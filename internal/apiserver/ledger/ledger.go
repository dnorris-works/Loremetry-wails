package ledger

import (
	"database/sql"
	"sync"
)

type Ledger struct {
	mu    sync.Mutex
	bal   map[string]int
	costs map[string]int
	db    *sql.DB
}

func New() *Ledger {
	return &Ledger{
		bal: map[string]int{
			"user_dev":   100,
			"user_empty": 0,
		},
		costs: map[string]int{},
	}
}

func NewPostgres(db *sql.DB) *Ledger {
	return &Ledger{db: db, costs: map[string]int{}}
}

func (l *Ledger) Balance(userID string) int {
	if l.db != nil {
		var n int
		_ = l.db.QueryRow(`SELECT balance FROM credit_ledger WHERE user_id = $1`, userID).Scan(&n)
		return n
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.bal[userID]
}

func (l *Ledger) CreditCost(analysisID string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n, ok := l.costs[analysisID]; ok {
		return n
	}
	return 1
}

func (l *Ledger) Debit(userID, analysisID string) (int, bool) {
	cost := l.CreditCost(analysisID)
	if l.db != nil {
		return l.debitDB(userID, analysisID, cost)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.bal[userID] < cost {
		return l.bal[userID], false
	}
	l.bal[userID] -= cost
	return l.bal[userID], true
}

func (l *Ledger) debitDB(userID, analysisID string, cost int) (int, bool) {
	tx, err := l.db.Begin()
	if err != nil {
		return 0, false
	}
	defer func() { _ = tx.Rollback() }()
	var bal int
	err = tx.QueryRow(`SELECT balance FROM credit_ledger WHERE user_id = $1 FOR UPDATE`, userID).Scan(&bal)
	if err != nil {
		_, _ = tx.Exec(`INSERT INTO credit_ledger (user_id, balance) VALUES ($1, 0) ON CONFLICT DO NOTHING`, userID)
		bal = 0
	}
	if bal < cost {
		return bal, false
	}
	bal -= cost
	_, err = tx.Exec(`UPDATE credit_ledger SET balance = $1 WHERE user_id = $2`, bal, userID)
	if err != nil {
		return 0, false
	}
	_, _ = tx.Exec(`INSERT INTO credit_events (user_id, analysis_id, delta, reason) VALUES ($1,$2,$3,'debit')`,
		userID, analysisID, -cost)
	if err := tx.Commit(); err != nil {
		return 0, false
	}
	return bal, true
}

func (l *Ledger) Add(userID string, n int) int {
	if l.db != nil {
		return l.addDB(userID, n, "grant")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bal[userID] += n
	return l.bal[userID]
}

func (l *Ledger) Refund(userID, analysisID string) int {
	cost := l.CreditCost(analysisID)
	if l.db != nil {
		return l.addDB(userID, cost, "refund")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bal[userID] += cost
	return l.bal[userID]
}

func (l *Ledger) addDB(userID string, n int, reason string) int {
	tx, err := l.db.Begin()
	if err != nil {
		return l.Balance(userID)
	}
	defer func() { _ = tx.Rollback() }()
	_, _ = tx.Exec(`INSERT INTO credit_ledger (user_id, balance) VALUES ($1, 0) ON CONFLICT DO NOTHING`, userID)
	var bal int
	_ = tx.QueryRow(`SELECT balance FROM credit_ledger WHERE user_id = $1 FOR UPDATE`, userID).Scan(&bal)
	bal += n
	_, _ = tx.Exec(`UPDATE credit_ledger SET balance = $1 WHERE user_id = $2`, bal, userID)
	_, _ = tx.Exec(`INSERT INTO credit_events (user_id, delta, reason) VALUES ($1,$2,$3)`, userID, n, reason)
	_ = tx.Commit()
	return bal
}
