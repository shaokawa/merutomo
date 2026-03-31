package auth

import (
	"net/mail"
	"strings"
)

const minPasswordLength = 8

type Service struct {
	supabase *SupabaseClient
}

type AuthResult struct {
	Token                  string
	User                   User
	NeedsEmailConfirmation bool
}

func NewService(supabase *SupabaseClient) *Service {
	return &Service{
		supabase: supabase,
	}
}

func (s *Service) Register(email, password string) (AuthResult, error) {
	if err := validateCredentials(email, password); err != nil {
		return AuthResult{}, err
	}

	resp, err := s.supabase.SignUp(email, password)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already registered") {
			return AuthResult{}, ErrEmailAlreadyExists
		}
		return AuthResult{}, err
	}

	if resp.User == nil {
		return AuthResult{}, ErrUnauthorized
	}

	result := AuthResult{
		User: User{
			ID:    resp.User.ID,
			Email: resp.User.Email,
		},
		NeedsEmailConfirmation: resp.Session == nil,
	}

	if resp.Session != nil {
		result.Token = resp.Session.AccessToken
	}

	return result, nil
}

func (s *Service) Login(email, password string) (AuthResult, error) {
	if err := validateCredentials(email, password); err != nil {
		return AuthResult{}, err
	}

	resp, err := s.supabase.SignInWithPassword(email, password)
	if err != nil {
		return AuthResult{}, err
	}

	if resp.User == nil {
		return AuthResult{}, ErrUnauthorized
	}

	return AuthResult{
		Token: resp.AccessToken,
		User: User{
			ID:    resp.User.ID,
			Email: resp.User.Email,
		},
	}, nil
}

func (s *Service) Authenticate(token string) (User, error) {
	user, err := s.supabase.GetUser(token)
	if err != nil {
		return User{}, err
	}
	return user, nil
}


func validateCredentials(email, password string) error {
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" {
		return ErrInvalidCredentials
	}

	if _, err := mail.ParseAddress(normalizedEmail); err != nil {
		return ErrInvalidCredentials
	}

	if len(password) < minPasswordLength {
		return ErrInvalidCredentials
	}

	return nil
}
