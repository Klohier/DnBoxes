package user

import (
	// "context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	// "github.com/stretchr/testify/suite"
)

type TestSuite struct {
	app *echo.Echo
	mockRepo UserRepository
	userHandler *UserHandler
	userService *UserService
}

func setup() *TestSuite {
	app := echo.New()
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)
	

	return &TestSuite{
		app: app,
		mockRepo: mockRepo,
		userHandler: userHandler,
		userService: userService,
	}
}


func TestCreateUser(t *testing.T){
	

	suite := setup()

	formData := "username=testing245&password=abcd"
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(formData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	ctx := suite.app.NewContext(req, rec)

	// Call the CreateUser method
	if err := suite.userHandler.CreateUser(ctx); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check response status and body
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	expectedResponse := UserResponse{
		UserID: 1,
		Username: "testing245",
	}

	jsonResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		t.Fatalf("Failed to generate expected JSON: %v", err)
	}
	// Using TrimSpace in order to compare the jsons as before test would fail even if content looked the same
	if strings.TrimSpace(rec.Body.String()) != strings.TrimSpace(string(jsonResponse)) {
		t.Errorf("Expected body %s, got %s", string(jsonResponse), rec.Body.String())

	}	
}

func TestFindByID(t *testing.T){
	// suite := setup()

}