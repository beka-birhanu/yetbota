package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/config"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmigrations"
	jwtDriver "github.com/beka-birhanu/yetbota/identity-service/drivers/jwt"
	logger "github.com/beka-birhanu/yetbota/identity-service/drivers/logger"
	neo4jDriver "github.com/beka-birhanu/yetbota/identity-service/drivers/neo4j"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/postgres"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/smstaskadapter"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/storage"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/endpoint"
	repoAuth "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/auth"
	repoFollow "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/follow"
	repoPhoto "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/photo"
	repoUser "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/user"
	usecaseAuth "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
	usecaseUser "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
	authGrpc "github.com/beka-birhanu/yetbota/identity-service/internal/transport/grpc/auth"
	userGrpc "github.com/beka-birhanu/yetbota/identity-service/internal/transport/grpc/user"
	"github.com/go-redis/redis/v8"
	"github.com/pressly/goose"
	"go.uber.org/zap/zapcore"

	cmdGrpc "github.com/beka-birhanu/yetbota/identity-service/cmd/grpc"
)

func main() {
	validator.InitValidator()
	logger.InitDefault(
		logger.MaskEnabled(),
		logger.WithStdout(),
		logger.WithLevel(zapcore.DebugLevel),
	)

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error load config: %v", err))
	}

	// Postgres
	pgdb, err := postgres.NewDB(
		&postgres.Config{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			User:     cfg.Postgres.User,
			Password: cfg.Postgres.Password,
			DB:       cfg.Postgres.DB,
		})
	if err != nil {
		panic(fmt.Errorf("error connect postgres: %v", err))
	}
	defer func() {
		_ = pgdb.Close()
	}()

	boil.SetDB(pgdb)

	if err := pgdb.Ping(); err != nil {
		panic(fmt.Errorf("error pinging database: %v", err))
	}
	fmt.Println("Database connection successful!")

	// Migrations
	dbGoose, err := dbmigrations.RunDBMigrations(&dbmigrations.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
	})
	if err != nil {
		panic(fmt.Errorf("error run DB migrations: %v", err))
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("error setting goose dialect: %v", err))
	}

	currentVersion, err := goose.GetDBVersion(dbGoose)
	if err != nil {
		fmt.Printf("Migration table initialization: %v\n", err)
	}
	fmt.Printf("Current migration version: %d\n", currentVersion)

	if err := goose.Up(dbGoose, constants.MigrationFolder); err != nil {
		panic(fmt.Errorf("error running migrations: %v", err))
	}

	// Redis
	redisConn := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
	})
	if _, err := redisConn.Ping(ctx).Result(); err != nil {
		panic(fmt.Errorf("error connecting to redis: %v", err))
	}
	fmt.Println("Redis connection successful!")

	// JWT Session Manager
	sessionManager, err := jwtDriver.NewSessionManager(&jwtDriver.Config{
		AccessKey:  cfg.Jwt.AccessToken.Secret,
		RefreshKey: cfg.Jwt.RefreshToken.Secret,
		AccessTTL:  time.Duration(cfg.Jwt.AccessToken.Expiration) * time.Second,
		RefreshTTL: time.Duration(cfg.Jwt.RefreshToken.Expiration) * time.Second,
		Algo:       cfg.Jwt.Algorithm,
		RedisConn:  redisConn,
	})
	if err != nil {
		panic(fmt.Errorf("error creating session manager: %v", err))
	}

	// Repositories
	userRepo, err := repoUser.NewRepo(&repoUser.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating user repo: %v", err))
	}

	otpStore, err := repoAuth.NewOtpStore(&repoAuth.OtpStoreConfig{RedisConn: redisConn})
	if err != nil {
		panic(fmt.Errorf("error creating otp store: %v", err))
	}

	// Hasher
	hasher := utils.NewHasher()

	// AWS S3
	awscfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWS.S3.Region))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}
	bucket, err := storage.NewS3Blob(&storage.S3Config{
		AwsConfig:         awscfg,
		DefaultBucketName: cfg.AWS.S3.Bucket,
	})
	if err != nil {
		panic(fmt.Errorf("error creating S3 blob: %v", err))
	}

	// Photo repo
	photoRepo, err := repoPhoto.NewRepo(&repoPhoto.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating photo repo: %v", err))
	}

	// Neo4j
	neo4jDrv, err := neo4jDriver.NewDriver(&neo4jDriver.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
	})
	if err != nil {
		panic(fmt.Errorf("error creating neo4j driver: %v", err))
	}
	defer func() {
		_ = neo4jDrv.Close(ctx)
	}()

	// Follow repo
	followRepo, err := repoFollow.NewRepo(&repoFollow.Config{Driver: neo4jDrv})
	if err != nil {
		panic(fmt.Errorf("error creating follow repo: %v", err))
	}

	// SMS Client
	smsClient := smstaskadapter.NewAdapter(&smstaskadapter.Config{})

	// Auth usecase
	authService, err := usecaseAuth.NewService(&usecaseAuth.Config{
		UserRepo:       userRepo,
		OtpStore:       otpStore,
		SessionManager: sessionManager,
		Hasher:         hasher,
		SMSClient:      smsClient,
		OtpTTL:         time.Duration(cfg.Otp.TTL) * time.Second,
		LockRequestTTL: time.Duration(cfg.Otp.LockRequestTime) * time.Second,
		LockInvalidTTL: time.Duration(cfg.Otp.LockInvalidTime) * time.Second,
		AccessTTL:      time.Duration(cfg.Jwt.AccessToken.Expiration) * time.Second,
		RefreshTTL:     time.Duration(cfg.Jwt.RefreshToken.Expiration) * time.Second,
	})
	if err != nil {
		panic(fmt.Errorf("error creating auth service: %v", err))
	}

	// User usecase
	userService, err := usecaseUser.NewService(&usecaseUser.Config{
		UserRepo:   userRepo,
		PhotoRepo:  photoRepo,
		FollowRepo: followRepo,
		OtpStore:   otpStore,
		Hasher:     hasher,
		Bucket:     bucket,
		BucketName: cfg.AWS.S3.Bucket,
	})
	if err != nil {
		panic(fmt.Errorf("error creating user service: %v", err))
	}

	// Endpoints
	endpoints, err := endpoint.NewEndpoints(&endpoint.Config{
		AuthService: authService,
		UserService: userService,
	})
	if err != nil {
		panic(fmt.Errorf("error creating endpoints: %v", err))
	}

	// gRPC Servers
	authServer, err := authGrpc.NewHandler(&authGrpc.Config{E: endpoints})
	if err != nil {
		panic(fmt.Errorf("error creating auth grpc server: %v", err))
	}
	userServer, err := userGrpc.NewHandler(&userGrpc.Config{E: endpoints})
	if err != nil {
		panic(fmt.Errorf("error creating user grpc server: %v", err))
	}

	// gRPC Server
	if err := cmdGrpc.RunServer(ctx, &cmdGrpc.Config{
		Port:                cfg.Grpc.Port,
		KeepaliveTime:       time.Duration(cfg.Grpc.Keepalive.Time) * time.Second,
		KeepaliveTimeout:    time.Duration(cfg.Grpc.Keepalive.Timeout) * time.Second,
		KeepaliveMinTime:    time.Duration(cfg.Grpc.Keepalive.MinTime) * time.Second,
		PermitWithoutStream: cfg.Grpc.Keepalive.PermitWithoutStream,
		SessionManager:      sessionManager,
		GrpcServers:         []cmdGrpc.GrpcServer{authServer, userServer},
	}); err != nil {
		panic(fmt.Errorf("gRPC server error: %v", err))
	}
}
