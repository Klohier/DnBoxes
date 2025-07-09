package auth

import (
	"context"
	"dango/internal/user"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	userRepo user.UserRepository
}

func NewLoginService(userRepo user.UserRepository) *LoginService {
	return &LoginService{
		userRepo: userRepo,
	}
}

func (s *LoginService) Login(ctx context.Context, username, password string) (*user.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}
