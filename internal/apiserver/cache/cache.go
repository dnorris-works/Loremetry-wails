package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"

	"loremetry/internal/apiserver/analysis"
)

type Store struct {
	mu    sync.Mutex
	items map[string]Entry
	db    *sql.DB
}

type Entry struct {
	AnalysisID string
	Body       string
}

func NewMemory() *Store {
	return &Store{items: map[string]Entry{}}
}

func NewPostgres(db *sql.DB) *Store {
	return &Store{db: db}
}

func SourcesHash(analysisID string, sources []analysis.Source) string {
	type part struct {
		key string
	}
	parts := make([]part, 0, len(sources))
	for _, s := range sources {
		textSum := sha256.Sum256([]byte(s.Text))
		parts = append(parts, part{
			key: fmt.Sprintf("%s:%s:%s", s.Role, s.Rel, hex.EncodeToString(textSum[:])),
		})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].key < parts[j].key })
	var b strings.Builder
	b.WriteString(analysisID)
	b.WriteByte('|')
	for i, p := range parts {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(p.key)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func (s *Store) Get(key string) (Entry, bool) {
	if s.db != nil {
		var e Entry
		err := s.db.QueryRow(`SELECT analysis_id, body FROM analysis_cache WHERE cache_key = $1`, key).
			Scan(&e.AnalysisID, &e.Body)
		if err != nil {
			return Entry{}, false
		}
		return e, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[key]
	return e, ok
}

func (s *Store) Put(key, analysisID, body string) {
	if s.db != nil {
		_, _ = s.db.Exec(`
			INSERT INTO analysis_cache (cache_key, analysis_id, body) VALUES ($1,$2,$3)
			ON CONFLICT (cache_key) DO UPDATE SET body = EXCLUDED.body, created_at = now()`,
			key, analysisID, body)
		return
	}
	s.mu.Lock()
	s.items[key] = Entry{AnalysisID: analysisID, Body: body}
	s.mu.Unlock()
}
