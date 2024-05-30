package threadpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var DefaultConfig = &Config{
	Concurrent:     100,
	IdleTimeout:    15 * time.Second,
	MaxWaitTimeout: 10 * time.Minute,
}

var pool *threadPool

type Config struct {
	Concurrent     int           `json:"concurrent" yaml:"concurrent"`
	IdleTimeout    time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxWaitTimeout time.Duration `json:"max_wait_timeout" yaml:"max_wait_timeout"`
	PoolName       string        `json:"pool_name" yaml:"pool_name"`
}

type funcWithCtx struct {
	f   func(ctx context.Context)
	ctx context.Context
}

type Pool struct {
	Cfg *Config

	// pool context
	oriCtx     context.Context
	poolCtx    context.Context
	poolCancel context.CancelFunc

	// pending is the chan to block Pool.Do() when go pool is full
	pending chan *funcWithCtx
	// workers is the counter of the worker
	workers chan struct{}

	mux            sync.RWMutex
	wg             sync.WaitGroup
	closed         bool
	aliveThreadNum int32
	idleThreadNum  int32
	name           string
}

func GetThreadPool() *threadPool {
	return pool
}

func InitThreadPool(cfg *Config) {
	pool = createThreadPool(context.Background(), cfg)
}

func CreateNewPool(ctx context.Context, name string, cfgs ...*Config) *Pool {
	cctx, cancel := context.WithCancel(ctx)
	if len(cfgs) == 0 {
		cfgs = append(cfgs, DefaultConfig)
	}

	if len(cfgs) == 0 {
		cfgs = append(cfgs, DefaultConfig)
	}

	cfg := cfgs[0]
	gr := &Pool{
		Cfg:        cfg,
		oriCtx:     ctx,
		poolCtx:    cctx,
		poolCancel: cancel,
		pending:    make(chan *funcWithCtx),
		workers:    make(chan struct{}, cfg.Concurrent),
		name:       name,
		mux:        sync.RWMutex{},
		wg:         sync.WaitGroup{},
	}

	return gr
}

func (g *Pool) Do(ctx context.Context, f func(ctx context.Context)) error {
	t := time.NewTimer(g.Cfg.MaxWaitTimeout)
	defer t.Stop()

	select {
	case g.pending <- &funcWithCtx{
		f:   f,
		ctx: ctx,
	}: // block if workers are busy
	case g.workers <- struct{}{}:
		g.wg.Add(1)
		atomic.AddInt32(&g.aliveThreadNum, 1)
		go g.loop(ctx, f)
	case <-ctx.Done():
		return fmt.Errorf("cancel by context")
	case <-t.C:
		return fmt.Errorf("get thread timeout")
	}
	return nil
}

func (g *Pool) loop(ctx context.Context, f func(ctx context.Context)) {
	defer g.wg.Done()
	defer atomic.AddInt32(&g.aliveThreadNum, -1)
	defer func() { <-g.workers }()

	timer := time.NewTimer(g.Cfg.IdleTimeout)
	defer timer.Stop()
	for {
		g.execute(ctx, f)
		atomic.AddInt32(&g.idleThreadNum, 1)
		select {
		case <-timer.C:
			atomic.AddInt32(&g.idleThreadNum, -1)
			return
		case stru := <-g.pending:
			atomic.AddInt32(&g.idleThreadNum, -1)
			if stru == nil {
				return
			}
			ctx = stru.ctx
			f = stru.f
			timer.Reset(g.Cfg.IdleTimeout)
		}
	}
}

func (g *Pool) execute(ctx context.Context, f func(ctx context.Context)) {
	f(ctx)
}

// Close will call context.Cancel(), so all goroutines maybe exit when job does not complete
func (g *Pool) Close(grace bool) {
	g.mux.Lock()
	if g.closed {
		g.poolCancel()
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pending)
	close(g.workers)

	g.poolCancel()
	if grace {
		g.wg.Wait()
	}
}

// Done will wait for all goroutines complete the jobs and then close the pool
func (g *Pool) Done() {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pending)
	close(g.workers)
	g.wg.Wait()
	g.poolCancel()
}

// Done will wait for all goroutines complete the jobs and then close the pool

type threadPool struct {
	*Pool
}

func (g *threadPool) OnConfigUpdate(cfg *Config) {
	oriPool := g.Pool
	g.Pool = CreateNewPool(g.oriCtx, cfg.PoolName, cfg)
	go func() {
		oriPool.wg.Wait()
		oriPool.poolCancel()
		oriPool.Done()
	}()

	return
}

func (g *threadPool) GetName() string {
	return g.Cfg.PoolName
}

func createThreadPool(ctx context.Context, cfg *Config) *threadPool {
	p := &threadPool{}
	p.Pool = CreateNewPool(ctx, cfg.PoolName, cfg)
	return p
}
