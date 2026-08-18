package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

type User struct {
	ID    string
	Email string
	Plan  string
}

type Store struct {
	mu     sync.Mutex
	tokens map[string]User
}

func New() *Store {
	s := &Store{tokens: map[string]User{}}
	s.tokens["dev-token"] = User{ID: "user_dev", Email: "dev@loremetry.local", Plan: "writer"}
	s.tokens["empty-token"] = User{ID: "user_empty", Email: "empty@loremetry.local", Plan: "writer"}
	return s
}

func (s *Store) Lookup(token string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.tokens[strings.TrimSpace(token)]
	return u, ok
}

func (s *Store) Issue(email string) (string, User) {
	email = strings.TrimSpace(email)
	if email == "" {
		email = "writer@loremetry.local"
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	u := User{ID: "user_" + token[:8], Email: email, Plan: "writer"}
	s.mu.Lock()
	s.tokens[token] = u
	s.mu.Unlock()
	return token, u
}

func Bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}
