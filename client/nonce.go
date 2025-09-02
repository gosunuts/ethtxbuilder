package client

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gosunuts/ethtxbuilder/utils"
	"github.com/umbracle/ethgo"
)

type NonceManager struct {
	c *Client

	mu    sync.Mutex
	cache map[string]*entry
	ttl   time.Duration
}

type entry struct {
	nonce       uint64
	lastSyncAt  time.Time
	lastBlockNo uint64
}

func NewNonceManager(c *Client, ttl time.Duration) *NonceManager {
	if ttl <= 0 {
		ttl = time.Minute
	}

	return &NonceManager{
		c:     c,
		cache: make(map[string]*entry),
		ttl:   ttl,
	}
}

func (m *NonceManager) Next(addr string) (uint64, error) {
	m.mu.Lock()
	e := m.cache[addr]
	needSync := e == nil || (m.ttl > 0 && time.Since(e.lastSyncAt) > m.ttl)
	m.mu.Unlock()

	if needSync {
		if err := m.Sync(addr); err != nil {
			return 0, err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	e = m.must(addr)
	n := e.nonce
	e.nonce++
	return n, nil
}

func (m *NonceManager) Sync(addr string) error {
	pendingNonce, err := m.pendingNonce(addr)
	if err != nil {
		return err
	}
	latestBlock, _ := m.c.BlockNumber()

	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.cache[addr]
	if e == nil {
		e = &entry{}
		m.cache[addr] = e
	}
	e.nonce = pendingNonce
	e.lastSyncAt = time.Now()
	e.lastBlockNo = latestBlock
	return nil
}

func (m *NonceManager) OnSendError(addr string, err error) (refreshed bool, _ error) {
	if err == nil {
		return false, nil
	}
	msg := strings.ToLower(err.Error())

	tooLow := strings.Contains(msg, "nonce too low") ||
		strings.Contains(msg, "already known") ||
		strings.Contains(msg, "replacement transaction underpriced")
	tooHigh := strings.Contains(msg, "nonce too high") ||
		strings.Contains(msg, "transaction nonce is too high")

	if tooLow || tooHigh {
		return true, m.Sync(addr)
	}
	return false, nil
}

// GetCached returns current cached next-nonce (best-effort).
func (m *NonceManager) GetCached(addr string) (uint64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.cache[addr]
	if e == nil {
		return 0, false
	}
	return e.nonce, true
}

// ---- Internals --------------------------------------------------------------

func (m *NonceManager) must(addr string) *entry {
	e := m.cache[addr]
	if e == nil {
		e = &entry{}
		m.cache[addr] = e
	}
	return e
}

// pendingNonce tries the typed call first; falls back to raw RPC if needed.
func (m *NonceManager) pendingNonce(addr string) (uint64, error) {
	if n, err := m.c.NonceAt(addr, ethgo.BlockNumber(ethgo.Pending)); err == nil {
		return n, nil
	}

	// 2) raw RPC fallback
	var out string
	if err := m.c.rpc.Call("eth_getTransactionCount", &out, addr, "pending"); err != nil {
		return 0, fmt.Errorf("pending nonce fetch failed: %w", err)
	}
	return utils.StrToU64(out)
}
