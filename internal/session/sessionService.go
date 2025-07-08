package session

import (
	"context"
	"fmt"
)

type SessionService struct {
	sessionRepo SessionRepository
}

func NewSessionService(sessionRepo SessionRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *SessionService) CreateSession(ctx context.Context) (*Session, error) {
	createdSession, err := s.sessionRepo.Create(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return createdSession, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID int) error {
	err := s.sessionRepo.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *SessionService) AddUserToSession(ctx context.Context, sessionID, userID int) error {
	return s.sessionRepo.AddUserToSession(ctx, sessionID, userID)
}

func (s *SessionService) RemoveUserFromSession(ctx context.Context, sessionID, userID int) error {
	return s.sessionRepo.RemoveUserFromSession(ctx, sessionID, userID)
}

func (s *SessionService) FindAllSessions(ctx context.Context) ([]Session, error) {
	sessions, err := s.sessionRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}
	return sessions, nil
}

func (s *SessionService) FindSessionByID(ctx context.Context, sessionID int) (*FullSession, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session by ID: %w", err)
	}
	return session, nil
}

func (s *SessionService) FindSessionByUserID(ctx context.Context, userID int) (*int, error) {
	sessionID, err := s.sessionRepo.FindSessionByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session for user %d: %w", userID, err)
	}
	return sessionID, nil
}

func (s *SessionService) SetUserConnectionStatus(ctx context.Context, sessionID, userID int, status string) error {
	return s.sessionRepo.SetUserConnectionStatus(ctx, sessionID, userID, status)
}
