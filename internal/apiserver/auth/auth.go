package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

type User struct {
	ID              string
	Email           string
	Plan            string
	MaxConcurrentAI int
}

type Store struct {
	mu     sync.Mutex
	tokens map[string]User
	db     *sql.DB
}

func New() *Store {
	s := &Store{tokens: map[string]User{}}
	s.tokens["dev-token"] = User{ID: "user_dev", Email: "dev@loremetry.local", Plan: "writer", MaxConcurrentAI: 1}
	s.tokens["empty-token"] = User{ID: "user_empty", Email: "empty@loremetry.local", Plan: "writer", MaxConcurrentAI: 1}
	return s
}

func NewPostgres(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Lookup(token string) (User, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, false
	}
	if s.db != nil {
		return s.lookupDB(token)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.tokens[token]
	return u, ok
}

func (s *Store) lookupDB(token string) (User, bool) {
	hash := HashToken(token)
	var u User
	err := s.db.QueryRow(`
		SELECT u.id, u.email, u.plan, u.max_concurrent_ai
		FROM device_tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.token_hash = $1`, hash).Scan(&u.ID, &u.Email, &u.Plan, &u.MaxConcurrentAI)
	if err != nil {
		return User{}, false
	}
	if u.MaxConcurrentAI <= 0 && u.Plan != "" && u.Plan != "none" {
		u.MaxConcurrentAI = planConcurrentDefault(u.Plan)
	}
	return u, true
}

func (s *Store) Issue(email string) (string, User) {
	email = strings.TrimSpace(email)
	if email == "" {
		email = "writer@loremetry.local"
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	u := User{
		ID:              "user_" + token[:8],
		Email:           email,
		Plan:            "writer",
		MaxConcurrentAI: 1,
	}
	if s.db != nil {
		s.issueDB(token, u)
		return token, u
	}
	s.mu.Lock()
	s.tokens[token] = u
	s.mu.Unlock()
	return token, u
}

func (s *Store) issueDB(token string, u User) {
	_, _ = s.db.Exec(`INSERT INTO users (id, email, plan, max_concurrent_ai) VALUES ($1,$2,$3,$4)
		ON CONFLICT (id) DO NOTHING`, u.ID, u.Email, u.Plan, u.MaxConcurrentAI)
	_, _ = s.db.Exec(`INSERT INTO credit_ledger (user_id, balance) VALUES ($1, 100)
		ON CONFLICT (user_id) DO NOTHING`, u.ID)
	_, _ = s.db.Exec(`INSERT INTO device_tokens (token_hash, user_id) VALUES ($1,$2)
		ON CONFLICT (token_hash) DO NOTHING`, HashToken(token), u.ID)
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func planConcurrentDefault(plan string) int {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "pro":
		return 3
	case "writer":
		return 1
	default:
		return 0
	}
}

func Bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}
