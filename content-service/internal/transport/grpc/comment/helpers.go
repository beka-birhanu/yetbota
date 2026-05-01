package comment

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/comment/v1"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func deadlineExceeded(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return status.Error(codes.Canceled, "The client canceled the request!")
	}
	return nil
}

func commentToProto(c *dbmodels.Comment) *pb.Comment {
	return &pb.Comment{
		Id:        c.ID,
		Comment:   c.Content,
		Upvote:    int32(c.Upvote),
		Downvote:  int32(c.Downvote),
		UserId:    c.UserID,
		PostId:    c.PostID,
		IsAnswer:  c.IsAnswer,
		CommentId: c.CommentID.String,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func mapCommentVoteTypeFromProto(v pb.CommentVoteType) string {
	switch v {
	case pb.CommentVoteType_COMMENT_VOTE_TYPE_UP:
		return dbmodels.CommentVoteTypeUpvote
	case pb.CommentVoteType_COMMENT_VOTE_TYPE_DOWN:
		return dbmodels.CommentVoteTypeDownvote
	default:
		return dbmodels.CommentVoteTypeUpvote
	}
}
