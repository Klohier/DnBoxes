package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo UserRepository

}

func NewUserService(userRepo UserRepository) *UserService{
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, username string, password string) (*User, error){
	if username == "" || password == "" {
		return nil, errors.New("username or password cannot be empty")
	}

	taken, err := s.userRepo.UserExists(ctx, username)
    if err != nil {
        return nil, err
    }
    if taken {
        return nil, errors.New("username is taken")
    }

	// Hashes Password From User
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err !=nil {
		return nil, err
	}
	

	newUser, err := s.userRepo.Create(ctx, username, string(hashedPassword))
	if err != nil {
        return nil, errors.New("Failed to Create User" + err.Error())
    }
	return newUser, nil
}

