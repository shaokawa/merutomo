package auth

import (
	"net/mail"
	"strings"
)

const minPasswordLength = 8

type Service struct {
	supabase     *SupabaseClient
	profileStore ProfileStore
}

type AuthResult struct {
	Token                  string
	User                   User
	NeedsEmailConfirmation bool
}

func NewService(supabase *SupabaseClient, profileStore ProfileStore) *Service {
	return &Service{
		supabase:     supabase,
		profileStore: profileStore,
	}
}

func (s *Service) Register(email, password, displayName, username, emailVisibility string) (AuthResult, error) {
	if err := validateCredentials(email, password); err != nil {
		return AuthResult{}, err
	}
	if err := validateRegistrationProfile(displayName, username, emailVisibility); err != nil {
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

	user, err := s.profileStore.CreateProfile(CreateProfileParams{
		ID:              resp.User.ID,
		Email:           resp.User.Email,
		DisplayName:     displayName,
		Username:        username,
		EmailVisibility: emailVisibility,
	})
	if err != nil {
		return AuthResult{}, err
	}

	result := AuthResult{
		User:                   user,
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

	user, err := s.profileStore.FindProfileByID(resp.User.ID)
	if err != nil {
		return AuthResult{}, err
	}
	user.Email = resp.User.Email

	return AuthResult{
		Token: resp.AccessToken,
		User:  user,
	}, nil
}

func (s *Service) Logout(token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrUnauthorized
	}

	return s.supabase.SignOut(token)
}

func (s *Service) Authenticate(token string) (User, error) {
	user, err := s.supabase.GetUser(token)
	if err != nil {
		return User{}, err
	}

	profile, err := s.profileStore.FindProfileByID(user.ID)
	if err != nil {
		return User{}, err
	}

	profile.Email = user.Email
	return profile, nil
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
