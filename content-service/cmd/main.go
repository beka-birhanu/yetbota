package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/beka-birhanu/yetbota/content-service/drivers/config"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmigrations"
	jwtDriver "github.com/beka-birhanu/yetbota/content-service/drivers/jwt"
	logger "github.com/beka-birhanu/yetbota/content-service/drivers/logger"
	"github.com/beka-birhanu/yetbota/content-service/drivers/postgres"
	"github.com/beka-birhanu/yetbota/content-service/drivers/storage"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/go-redis/redis/v8"
	"github.com/pressly/goose"
	"go.uber.org/zap/zapcore"

	cmdGrpc "github.com/beka-birhanu/yetbota/content-service/cmd/grpc"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	repoPhoto "github.com/beka-birhanu/yetbota/content-service/internal/services/repository/photo"
	repoPost "github.com/beka-birhanu/yetbota/content-service/internal/services/repository/post"
	repoComment "github.com/beka-birhanu/yetbota/content-service/internal/services/repository/comment"
	repoPostPhoto "github.com/beka-birhanu/yetbota/content-service/internal/services/repository/postphoto"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
	grpcComment "github.com/beka-birhanu/yetbota/content-service/internal/transport/grpc/comment"
	grpcPost "github.com/beka-birhanu/yetbota/content-service/internal/transport/grpc/post"
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

	pgdb, err := postgres.NewDB(&postgres.Config{
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

	redisConn := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
	})
	if _, err := redisConn.Ping(ctx).Result(); err != nil {
		panic(fmt.Errorf("error connecting to redis: %v", err))
	}
	fmt.Println("Redis connection successful!")

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

	postRepo, err := repoPost.NewRepo(&repoPost.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating post repo: %v", err))
	}

	photoRepo, err := repoPhoto.NewRepo(&repoPhoto.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating photo repo: %v", err))
	}

	postPhotoRepo, err := repoPostPhoto.NewRepo(&repoPostPhoto.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating postphoto repo: %v", err))
	}

	awscfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWS.S3.Region))
	if err != nil {
		panic(fmt.Errorf("error loading AWS config: %v", err))
	}
	bucket, err := storage.NewS3Blob(&storage.S3Config{
		AwsConfig:         awscfg,
		DefaultBucketName: cfg.AWS.S3.Bucket,
	})
	if err != nil {
		panic(fmt.Errorf("error creating S3 blob: %v", err))
	}

	postService, err := postSvc.NewService(&postSvc.Config{
		PostRepo:      postRepo,
		PhotoRepo:     photoRepo,
		PostPhotoRepo: postPhotoRepo,
		Bucket:        bucket,
		BucketName:    cfg.AWS.S3.Bucket,
	})
	if err != nil {
		panic(fmt.Errorf("error creating post service: %v", err))
	}

	commentRepo, err := repoComment.NewRepo(&repoComment.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating comment repo: %v", err))
	}

	commentService, err := commentSvc.NewService(&commentSvc.Config{
		CommentRepo: commentRepo,
	})
	if err != nil {
		panic(fmt.Errorf("error creating comment service: %v", err))
	}

	endpoints, err := endpoint.NewEndpoints(&endpoint.Config{
		PostService:    postService,
		CommentService: commentService,
	})
	if err != nil {
		panic(fmt.Errorf("error creating endpoints: %v", err))
	}

	postHandler, err := grpcPost.NewHandler(&grpcPost.Config{E: endpoints})
	if err != nil {
		panic(fmt.Errorf("error creating post handler: %v", err))
	}

	commentHandler, err := grpcComment.NewHandler(&grpcComment.Config{E: endpoints})
	if err != nil {
		panic(fmt.Errorf("error creating comment handler: %v", err))
	}

	if err := cmdGrpc.RunServer(ctx, &cmdGrpc.Config{
		Port:                cfg.Grpc.Port,
		KeepaliveTime:       time.Duration(cfg.Grpc.Keepalive.Time) * time.Second,
		KeepaliveTimeout:    time.Duration(cfg.Grpc.Keepalive.Timeout) * time.Second,
		KeepaliveMinTime:    time.Duration(cfg.Grpc.Keepalive.MinTime) * time.Second,
		PermitWithoutStream: cfg.Grpc.Keepalive.PermitWithoutStream,
		SessionManager:      sessionManager,
		GrpcServers:         []cmdGrpc.GrpcServer{postHandler, commentHandler},
	}); err != nil {
		panic(fmt.Errorf("gRPC server error: %v", err))
	}
}
