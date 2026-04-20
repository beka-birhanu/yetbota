package user

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
)

func resolutionFromPb(r pb.PhotoResolution) userSvc.PhotoResolution {
	switch r {
	case pb.PhotoResolution_PHOTO_RESOLUTION_MOBILE:
		return userSvc.PhotoResolutionMobile
	case pb.PhotoResolution_PHOTO_RESOLUTION_WEB:
		return userSvc.PhotoResolutionWeb
	case pb.PhotoResolution_PHOTO_RESOLUTION_ORIGINAL:
		return userSvc.PhotoResolutionOriginal
	default:
		return userSvc.PhotoResolutionUnspecified
	}
}

func roleFromPb(s pb.Role) string {
	switch s {
	case pb.Role_ROLE_USER:
		return dbmodels.RolesUSER
	case pb.Role_ROLE_ADMIN:
		return dbmodels.RolesADMIN
	default:
		return ""
	}
}

func roleToProto(s string) pb.Role {
	switch s {
	case dbmodels.RolesUSER:
		return pb.Role_ROLE_USER
	case dbmodels.RolesADMIN:
		return pb.Role_ROLE_ADMIN
	default:
		return pb.Role_ROLE_UNSPECIFIED
	}
}

func userToPrivateUser(u *domainUser.User, profileURL string) *pb.PrivateUser {
	badges := make([]pb.Badge, 0, len(u.Badges))
	// Badges stored as strings; map to proto enum if applicable
	_ = u.Badges

	return &pb.PrivateUser{
		Id:            u.ID,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Username:      u.Username,
		Mobile:        u.Mobile,
		Rating:        int32(u.Rating),
		Badges:        badges,
		Contributions: int32(u.Contributions),
		Followers:     int32(u.Followers),
		Following:     int32(u.Following),
		Status:        u.Status,
		Role:          roleToProto(u.Role),
		ProfileUrl:    profileURL,
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
	}
}

func userToPublicUser(u *domainUser.User, profileURL string) *pb.PublicUser {
	return &pb.PublicUser{
		Id:             u.ID,
		Username:       u.Username,
		MobileVerified: u.Mobile != "",
		Rating:         int32(u.Rating),
		Badges:         []pb.Badge{},
		Contributions:  int32(u.Contributions),
		Followers:      int32(u.Followers),
		Following:      int32(u.Following),
		ProfileUrl:     profileURL,
		CreatedAt:      timestamppb.New(u.CreatedAt),
	}
}

func listResponseToProto(r *userSvc.ListResponse) *pb.UserList {
	users := make([]*pb.PrivateUser, 0, len(r.Users))
	for _, w := range r.Users {
		users = append(users, userToPrivateUser(w.User, w.ProfileURL))
	}
	return &pb.UserList{
		Users: users,
		Pagination: &pb.PaginationInfo{
			Total:       r.Pagination.Total,
			Limit:       int32(r.Pagination.Limit),
			CurrentPage: int32(r.Pagination.CurrentPage),
		},
	}
}

func readResponseToProto(r *userSvc.ReadResponse) *pb.UserReadData {
	if r.UserWrapper == nil || r.User == nil {
		return &pb.UserReadData{}
	}
	return &pb.UserReadData{
		User:       userToPrivateUser(r.User, r.ProfileURL),
		ProfileUrl: r.ProfileURL,
	}
}

func readPublicResponseToProto(r *userSvc.ReadPublicResponse) *pb.PublicUser {
	if r.UserWrapper == nil || r.User == nil {
		return nil
	}
	return userToPublicUser(r.User, r.ProfileURL)
}

func mapSortField(s pb.SortField) domainUser.SortField {
	switch s {
	case pb.SortField_SORT_FIELD_RATING:
		return domainUser.SortFieldRating
	case pb.SortField_SORT_FIELD_FOLLOWERS:
		return domainUser.SortFieldFollowers
	case pb.SortField_SORT_FIELD_FOLLOWING:
		return domainUser.SortFieldFollowing
	case pb.SortField_SORT_FIELD_CONTRIBUTIONS:
		return domainUser.SortFieldContributions
	case pb.SortField_SORT_FIELD_CREATED_AT:
		return domainUser.SortFieldCreatedAt
	default:
		return ""
	}
}

func mapSortDirection(s pb.SortDirection) domainUser.SortDirection {
	switch s {
	case pb.SortDirection_SORT_DIRECTION_ASC:
		return domainUser.SortDirectionAsc
	case pb.SortDirection_SORT_DIRECTION_DESC:
		return domainUser.SortDirectionDesc
	default:
		return ""
	}
}
