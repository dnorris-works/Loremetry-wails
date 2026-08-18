package ledger

import "sync"

type Ledger struct {
	mu    sync.Mutex
	bal   map[string]int
	costs map[string]int
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

func (l *Ledger) Balance(userID string) int {
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
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.bal[userID] < cost {
		return l.bal[userID], false
	}
	l.bal[userID] -= cost
	return l.bal[userID], true
}

func (l *Ledger) Add(userID string, n int) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bal[userID] += n
	return l.bal[userID]
}
