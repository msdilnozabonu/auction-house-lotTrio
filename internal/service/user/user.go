package user

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/user"
	"context"
)

type Service interface {
	Me(ctx context.Context, userID int64) (model.User, error)
}

type service struct {
	userRepo user.Repo
}

func New(userRepo user.Repo) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Me(ctx context.Context, userID int64) (model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
