package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return errors.New("username must be between 3 and 20 characters")
	}
	if !validUsername.MatchString(username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return errors.New("password must be at most 72 characters")
	}
	return nil
}

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) FindByID(ctx context.Context, userId int) (*User, error) {
	user, err := s.userRepo.FindByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user, nil

}

func (s *UserService) CreateUser(ctx context.Context, username string, password string) (*User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}

	newUser, err := s.userRepo.Create(ctx, username, string(hashedPassword))
	if err != nil {
		return nil, errors.New("Failed to Create User")
	}
	return newUser, nil
}

func (s *UserService) CreateGuestUser(ctx context.Context) (*User, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return nil, errors.New("failed to generate guest username")
	}
	username := "Guest_" + hex.EncodeToString(suffix)

	guest, err := s.userRepo.CreateGuest(ctx, username)
	if err != nil {
		return nil, errors.New("failed to create guest user: " )
	}
	return guest, nil
}

func (s *UserService) UpgradeGuest(ctx context.Context, userID int, username string, password string) (*User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	taken, err := s.userRepo.UserExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, errors.New("username is taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.userRepo.UpgradeGuest(ctx, userID, username, string(hashedPassword))
}

func (s *UserService) UpdateGameID(ctx context.Context, userId int, gameId *int) (*User, error) {
	user, err := s.userRepo.UpdateGameID(ctx, userId, gameId)
	if err != nil {
		return nil, err
	}

	return user, nil
}
