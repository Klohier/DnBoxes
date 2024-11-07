package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

type MockUserRepository struct {
	users map[int]*User
	usernames map[string]int
	nextID int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[int]*User),
		usernames: make(map[string]int),
		nextID:    0,
	}
}


func (m *MockUserRepository) FindAll(ctx context.Context) ([]User, error){
	return nil, nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int) (*User, error){
	if user, exists := m.users[id]; exists{
		return user, nil
	}
	return nil, fmt.Errorf("user with id %d not found", id) 
}
	

	
func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*User, error){
	return nil, nil
}


func (m *MockUserRepository) Create(ctx context.Context, username string, password string) (*User, error){
	if id, exists := m.usernames[username]; exists {
		return nil, errors.New("username already taken" + strconv.Itoa(id))
	}

	m.nextID++
	
	newUser := &User{UserID: m.nextID, Username: username, Password: password}
	m.users[m.nextID] = newUser
	m.usernames[username] = m.nextID
	return newUser, nil
}

func (m *MockUserRepository) UserExists(ctx context.Context, username string) (bool, error) {
    user, err := m.FindByUsername(ctx, username)
    if err != nil {
        return false, err
    }
    return user != nil, nil
}