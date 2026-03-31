package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SupabaseClient struct {
	baseURL    string
	anonKey    string
	httpClient *http.Client
}

type supabaseSignUpRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Options  *supabaseSignUpOptions `json:"options,omitempty"`
}

type supabaseSignUpOptions struct {
	Data map[string]any `json:"data,omitempty"`
}

type supabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type supabaseSession struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type supabaseSignUpResponse struct {
	User    *supabaseUser    `json:"user"`
	Session *supabaseSession `json:"session"`
}

type supabaseErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type SupabaseAPIError struct {
	Status int
	Body   string
}

type supabaseTokenRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

type supabaseAuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	User         *supabaseUser `json:"user"`
}


func (e *SupabaseAPIError) Error() string {
	return fmt.Sprintf("signup failed: status=%d body=%s", e.Status, e.Body)
}
func NewSupabaseClient(baseURL, anonKey string) *SupabaseClient {
	return &SupabaseClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		anonKey: anonKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *SupabaseClient) SignUp(email, password string) (supabaseSignUpResponse, error) {
	payload := supabaseSignUpRequest{
		Email:    strings.TrimSpace(email),
		Password: password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return supabaseSignUpResponse{}, fmt.Errorf("marshal signup payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/v1/signup", bytes.NewReader(body))
	if err != nil {
		return supabaseSignUpResponse{}, fmt.Errorf("build signup request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return supabaseSignUpResponse{}, fmt.Errorf("send signup request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return supabaseSignUpResponse{}, fmt.Errorf("read signup response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result supabaseSignUpResponse
		if err := json.Unmarshal(raw, &result); err != nil {
			return supabaseSignUpResponse{}, fmt.Errorf("decode signup success response: %w; body=%s", err, string(raw))
		}
		return result, nil
	}

	return supabaseSignUpResponse{}, &SupabaseAPIError{
		Status: resp.StatusCode,
		Body:   strings.TrimSpace(string(raw)),
	}
}

func (c *SupabaseClient) SignInWithPassword(email, password string) (supabaseAuthResponse, error) {
	payload := supabaseTokenRequest{
		Email:    strings.TrimSpace(email),
		Password: password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return supabaseAuthResponse{}, fmt.Errorf("marshal login payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/v1/token?grant_type=password", bytes.NewReader(body))
	if err != nil {
		return supabaseAuthResponse{}, fmt.Errorf("build login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return supabaseAuthResponse{}, fmt.Errorf("send login request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return supabaseAuthResponse{}, fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result supabaseAuthResponse
		if err := json.Unmarshal(raw, &result); err != nil {
			return supabaseAuthResponse{}, fmt.Errorf("decode login response: %w; body=%s", err, string(raw))
		}
		return result, nil
	}

	return supabaseAuthResponse{}, &SupabaseAPIError{
		Status: resp.StatusCode,
		Body:   strings.TrimSpace(string(raw)),
	}
}

func (c *SupabaseClient) GetUser(accessToken string) (User, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/auth/v1/user", nil)
	if err != nil {
		return User{}, fmt.Errorf("build get user request: %w", err)
	}

	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("send get user request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, fmt.Errorf("read get user response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result supabaseUser
		if err := json.Unmarshal(raw, &result); err != nil {
			return User{}, fmt.Errorf("decode get user response: %w; body=%s", err, string(raw))
		}
		return User{
			ID:    result.ID,
			Email: result.Email,
		}, nil
	}

	return User{}, &SupabaseAPIError{
		Status: resp.StatusCode,
		Body:   strings.TrimSpace(string(raw)),
	}
}
