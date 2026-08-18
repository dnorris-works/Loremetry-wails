package ledger

import "testing"

func TestDebitEmpty(t *testing.T) {
	l := New()
	if _, ok := l.Debit("user_empty", "genre_analysis"); ok {
		t.Fatal("expected empty ledger to refuse debit")
	}
	left, ok := l.Debit("user_dev", "genre_analysis")
	if !ok || left != 99 {
		t.Fatalf("got %d %v", left, ok)
	}
}
