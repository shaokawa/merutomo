package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaokawa/merutomo/backend/internal/auth"
)

func TestAuthFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newTestRouter()

	registerBody := map[string]string{
		"email":            "user@example.com",
		"password":         "password123",
		"display_name":     "Test User",
		"username":         "test_user",
		"email_visibility": "approval_required",
	}

	registerResp := performJSONRequest(t, engine, http.MethodPost, "/auth/register", registerBody, "")
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", registerResp.Code)
	}

	var registerPayload struct {
		Token string `json:"token"`
		User  struct {
			Email           string `json:"email"`
			DisplayName     string `json:"display_name"`
			Username        string `json:"username"`
			EmailVisibility string `json:"email_visibility"`
		} `json:"user"`
	}
	if err := json.Unmarshal(registerResp.Body.Bytes(), &registerPayload); err != nil {
		t.Fatalf("failed to parse register response: %v", err)
	}

	if registerPayload.Token == "" {
		t.Fatal("expected token in register response")
	}

	if registerPayload.User.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", registerPayload.User.Email)
	}
	if registerPayload.User.DisplayName != "Test User" {
		t.Fatalf("expected display_name in register response, got %q", registerPayload.User.DisplayName)
	}
	if registerPayload.User.Username != "test_user" {
		t.Fatalf("expected username in register response, got %q", registerPayload.User.Username)
	}
	if registerPayload.User.EmailVisibility != "approval_required" {
		t.Fatalf("expected email_visibility in register response, got %q", registerPayload.User.EmailVisibility)
	}

	meResp := performJSONRequest(t, engine, http.MethodGet, "/auth/me", nil, registerPayload.Token)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected 200 from /auth/me, got %d", meResp.Code)
	}

	loginResp := performJSONRequest(t, engine, http.MethodPost, "/auth/login", registerBody, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d", loginResp.Code)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newTestRouter()

	body := map[string]string{
		"email":            "user@example.com",
		"password":         "password123",
		"display_name":     "Test User",
		"username":         "test_user",
		"email_visibility": "approval_required",
	}

	firstResp := performJSONRequest(t, engine, http.MethodPost, "/auth/register", body, "")
	if firstResp.Code != http.StatusCreated {
		t.Fatalf("expected first register request to succeed, got %d", firstResp.Code)
	}

	secondResp := performJSONRequest(t, engine, http.MethodPost, "/auth/register", body, "")
	if secondResp.Code != http.StatusConflict {
		t.Fatalf("expected duplicate register to return 409, got %d", secondResp.Code)
	}
}

func TestMeRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newTestRouter()

	resp := performJSONRequest(t, engine, http.MethodGet, "/auth/me", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.Code)
	}
}

func TestLogoutReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newTestRouter()

	resp := performJSONRequest(t, engine, http.MethodPost, "/auth/logout", nil, "token-value")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from logout, got %d", resp.Code)
	}
}

func newTestRouter() *gin.Engine {
	engine := gin.New()
	handler := auth.NewHandler(newFakeAuthService())
	Setup(engine, handler)

	return engine
}

type fakeAuthService struct {
	usersByEmail map[string]fakeAuthUser
	usersByToken map[string]fakeAuthUser
	nextID       int
}

type fakeAuthUser struct {
	ID              string
	Email           string
	Password        string
	Token           string
	DisplayName     string
	Username        string
	EmailVisibility string
}

func newFakeAuthService() *fakeAuthService {
	return &fakeAuthService{
		usersByEmail: make(map[string]fakeAuthUser),
		usersByToken: make(map[string]fakeAuthUser),
	}
}

func (s *fakeAuthService) Register(email, password, displayName, username, emailVisibility string) (auth.AuthResult, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	normalizedVisibility := strings.ToLower(strings.TrimSpace(emailVisibility))
	if normalizedVisibility == "" {
		normalizedVisibility = "approval_required"
	}
	if normalizedEmail == "" || len(password) < 8 || strings.TrimSpace(displayName) == "" || normalizedUsername == "" {
		return auth.AuthResult{}, auth.ErrInvalidProfile
	}
	switch normalizedVisibility {
	case "public", "approval_required", "private":
	default:
		return auth.AuthResult{}, auth.ErrInvalidProfile
	}

	if _, exists := s.usersByEmail[normalizedEmail]; exists {
		return auth.AuthResult{}, auth.ErrEmailAlreadyExists
	}

	s.nextID++
	user := fakeAuthUser{
		ID:              "user-id",
		Email:           normalizedEmail,
		Password:        password,
		Token:           "token-value",
		DisplayName:     strings.TrimSpace(displayName),
		Username:        normalizedUsername,
		EmailVisibility: normalizedVisibility,
	}

	if s.nextID > 1 {
		user.ID = "user-id-" + string(rune('0'+s.nextID))
		user.Token = "token-value-" + string(rune('0'+s.nextID))
	}

	s.usersByEmail[normalizedEmail] = user
	s.usersByToken[user.Token] = user

	return auth.AuthResult{
		Token: user.Token,
		User: auth.User{
			ID:              user.ID,
			Email:           user.Email,
			DisplayName:     user.DisplayName,
			Username:        user.Username,
			EmailVisibility: user.EmailVisibility,
		},
	}, nil
}

func (s *fakeAuthService) Login(email, password string) (auth.AuthResult, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	user, exists := s.usersByEmail[normalizedEmail]
	if !exists || user.Password != password {
		return auth.AuthResult{}, auth.ErrInvalidCredentials
	}

	return auth.AuthResult{
		Token: user.Token,
		User: auth.User{
			ID:              user.ID,
			Email:           user.Email,
			DisplayName:     user.DisplayName,
			Username:        user.Username,
			EmailVisibility: user.EmailVisibility,
		},
	}, nil
}

func (s *fakeAuthService) Logout(token string) error {
	if strings.TrimSpace(token) == "" {
		return auth.ErrUnauthorized
	}

	return nil
}

func (s *fakeAuthService) Authenticate(token string) (auth.User, error) {
	user, exists := s.usersByToken[token]
	if !exists {
		return auth.User{}, auth.ErrUnauthorized
	}

	return auth.User{
		ID:              user.ID,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		Username:        user.Username,
		EmailVisibility: user.EmailVisibility,
	}, nil
}

func performJSONRequest(t *testing.T, engine *gin.Engine, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	return recorder
}
