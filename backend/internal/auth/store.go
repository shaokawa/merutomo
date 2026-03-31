package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	CreatedAt    time.Time
}

type session struct {
	UserID    string
	ExpiresAt time.Time
}

type MemoryStore struct {
	mu           sync.RWMutex
	usersByID    map[string]User
	userIDsByKey map[string]string
	sessions     map[string]session
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		usersByID:    make(map[string]User),
		userIDsByKey: make(map[string]string),
		sessions:     make(map[string]session),
	}
}

func (s *MemoryStore) CreateUser(email string, passwordHash []byte) (User, error) {
	normalizedEmail := normalizeEmail(email)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.userIDsByKey[normalizedEmail]; exists {
		return User{}, ErrEmailAlreadyExists
	}

	user := User{
		ID:           mustGenerateToken(16),
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	s.usersByID[user.ID] = user
	s.userIDsByKey[normalizedEmail] = user.ID

	return user, nil
}

func (s *MemoryStore) FindUserByEmail(email string) (User, bool) {
	normalizedEmail := normalizeEmail(email)

	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.userIDsByKey[normalizedEmail]
	if !ok {
		return User{}, false
	}

	user, ok := s.usersByID[userID]
	return user, ok
}

func (s *MemoryStore) FindUserByID(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByID[id]
	return user, ok
}

func (s *MemoryStore) CreateSession(userID string, ttl time.Duration) (string, error) {
	token, err := generateToken(32)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = session{
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(ttl),
	}

	return token, nil
}

func (s *MemoryStore) FindUserBySession(token string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentSession, ok := s.sessions[token]
	if !ok {
		return User{}, false
	}

	if time.Now().UTC().After(currentSession.ExpiresAt) {
		delete(s.sessions, token)
		return User{}, false
	}

	user, ok := s.usersByID[currentSession.UserID]
	return user, ok
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func mustGenerateToken(size int) string {
	token, err := generateToken(size)
	if err != nil {
		panic(err)
	}

	return token
}

func generateToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
