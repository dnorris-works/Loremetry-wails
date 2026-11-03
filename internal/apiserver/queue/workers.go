package queue

import (
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Workers struct {
	svc        *Service
	minWorkers int
	maxWorkers int
	active     atomic.Int32
	stop       chan struct{}
	wg         sync.WaitGroup
}

func NewWorkers(svc *Service) *Workers {
	minW := envInt("LOREMETRY_AI_WORKERS_MIN", 1)
	maxW := envInt("LOREMETRY_AI_WORKERS_MAX", 4)
	if maxW < minW {
		maxW = minW
	}
	return &Workers{
		svc:        svc,
		minWorkers: minW,
		maxWorkers: maxW,
		stop:       make(chan struct{}),
	}
}

func (w *Workers) Start() {
	w.wg.Add(1)
	go w.loop()
}

func (w *Workers) Stop() {
	close(w.stop)
	w.wg.Wait()
}

func (w *Workers) loop() {
	defer w.wg.Done()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-tick.C:
			w.pump()
		}
	}
}

func (w *Workers) pump() {
	target := w.targetWorkers()
	for int(w.active.Load()) < target {
		j, err := w.svc.claimNext()
		if err != nil || j == nil {
			break
		}
		w.active.Add(1)
		go func(job *Job) {
			defer w.active.Add(-1)
			w.svc.runJob(job)
		}(j)
	}
}

func (w *Workers) targetWorkers() int {
	depth := w.queueDepth()
	n := (depth + 2) / 3
	if n < w.minWorkers {
		n = w.minWorkers
	}
	if n > w.maxWorkers {
		n = w.maxWorkers
	}
	return n
}

func (w *Workers) queueDepth() int {
	if w.svc.db != nil {
		var n int
		_ = w.svc.db.QueryRow(`SELECT count(*) FROM ai_jobs WHERE status = 'queued'`).Scan(&n)
		return n
	}
	w.svc.mu.Lock()
	defer w.svc.mu.Unlock()
	n := 0
	for _, j := range w.svc.jobs {
		if j.Status == StatusQueued {
			n++
		}
	}
	return n
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
