package gapi

import (
	db "github.com/KamisAyaka/simplebank/db/sqlc"
	"github.com/KamisAyaka/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func converterUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreateAt:          timestamppb.New(user.CreatedAt),
	}
}
