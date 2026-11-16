package health

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StatusOK       = "ok"
	StatusDegraded = "degraded"
	StatusFail     = "fail"
)

type CheckFunc func(context.Context) error

type Check struct {
	Name     string
	Critical bool
	Fn       CheckFunc
}

type Result struct {
	Status   string `json:"status"`
	Critical bool   `json:"critical"`
	Error    string `json:"error,omitempty"`
}

type Checker struct {
	checks  []Check
	timeout time.Duration
	ready   atomic.Bool
}

func StatusText(ok bool, results map[string]Result) string {
	if !ok {
		return StatusFail // упала критичная
	}
	for _, r := range results {
		if r.Status != "ok" {
			return StatusDegraded // упала некритичная
		}
	}
	return StatusOK
}

func New(timeout time.Duration, checks ...Check) *Checker {
	return &Checker{timeout: timeout, checks: checks}
}

func (c *Checker) SetReady(v bool) { c.ready.Store(v) }
func (c *Checker) Ready() bool     { return c.ready.Load() }

func (c *Checker) Run(ctx context.Context) (ok bool, results map[string]Result) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	results = make(map[string]Result, len(c.checks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, chk := range c.checks {
		wg.Add(1)
		go func(chk Check) {
			defer wg.Done()
			r := Result{Status: "ok", Critical: chk.Critical}
			if err := chk.Fn(ctx); err != nil {
				r.Status, r.Error = "fail", err.Error()
			}
			mu.Lock()
			results[chk.Name] = r
			mu.Unlock()
		}(chk)
	}
	wg.Wait()

	ok = true
	for _, r := range results {
		if r.Critical && r.Status != "ok" {
			ok = false
		}
	}
	return ok, results
}
