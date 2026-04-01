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
	Register(email, password, displayName, username, emailVisibility string) (AuthResult, error)
	Login(email, password string) (AuthResult, error)
	Logout(token string) error
	Authenticate(token string) (User, error)
}

type userResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	DisplayName     string `json:"display_name"`
	Username        string `json:"username"`
	ProfileText     string `json:"profile_text"`
	AvatarURL       string `json:"avatar_url"`
	EmailVisibility string `json:"email_visibility"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	DisplayName     string `json:"display_name"`
	Username        string `json:"username"`
	EmailVisibility string `json:"email_visibility"`
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewHandler(service authService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.Register(req.Email, req.Password, req.DisplayName, req.Username, req.EmailVisibility)
	if err != nil {
		var apiErr *SupabaseAPIError

		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case errors.Is(err, ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is invalid"})
		case errors.Is(err, ErrInvalidProfile):
			c.JSON(http.StatusBadRequest, gin.H{"error": "display_name, username, or email_visibility is invalid"})
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
			var apiErr *SupabaseAPIError

			switch {
			case errors.Is(err, ErrUnauthorized):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case errors.Is(err, ErrUserProfileNotFound):
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user profile not found"})
			case errors.As(err, &apiErr):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
			}
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
		ID:              user.ID,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		Username:        user.Username,
		ProfileText:     user.ProfileText,
		AvatarURL:       user.AvatarURL,
		EmailVisibility: user.EmailVisibility,
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
