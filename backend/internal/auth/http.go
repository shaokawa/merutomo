package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const userContextKey = "auth_user"

type Handler struct {
	service authService
}

type authService interface {
	Register(email, password string) (AuthResult, error)
	Login(email, password string) (AuthResult, error)
	Logout(token string) error
	Authenticate(token string) (User, error)
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewHandler(service authService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		var apiErr *SupabaseAPIError

		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is invalid"})
		case errors.As(err, &apiErr):
			c.JSON(apiErr.Status, gin.H{
				"error":   "supabase auth error",
				"details": apiErr.Body,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to register user",
				"details": err.Error(),
			})
		}
		return
	}

	status := http.StatusCreated
	if result.NeedsEmailConfirmation {
		c.JSON(status, gin.H{
			"user":                     toUserResponse(result.User),
			"token":                    result.Token,
			"needs_email_confirmation": true,
		})
		return
	}

	c.JSON(status, gin.H{
		"user":                     toUserResponse(result.User),
		"token":                    result.Token,
		"needs_email_confirmation": false,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		var apiErr *SupabaseAPIError

		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case errors.As(err, &apiErr):
			c.JSON(apiErr.Status, gin.H{
				"error":   "supabase auth error",
				"details": apiErr.Body,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to login",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, buildAuthResponse(result))
}

func (h *Handler) Logout(c *gin.Context) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return
	}

	err := h.service.Logout(token)
	if err != nil {
		var apiErr *SupabaseAPIError

		switch {
		case errors.As(err, &apiErr):
			c.JSON(apiErr.Status, gin.H{
				"error":   "supabase auth error",
				"details": apiErr.Body,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to logout",
				"details": err.Error(),
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}


func (h *Handler) Me(c *gin.Context) {
	user, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toUserResponse(user),
	})
}

func (h *Handler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		user, err := h.service.Authenticate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(userContextKey, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (User, bool) {
	rawUser, ok := c.Get(userContextKey)
	if !ok {
		return User{}, false
	}

	user, ok := rawUser.(User)
	return user, ok
}

func buildAuthResponse(result AuthResult) authResponse {
	return authResponse{
		Token: result.Token,
		User:  toUserResponse(result.User),
	}
}

func toUserResponse(user User) userResponse {
	return userResponse{
		ID:    user.ID,
		Email: user.Email,
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
