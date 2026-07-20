package security

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"go-crud-db-p2/config"
)

type MemoryStateStore struct {
	mu        sync.Mutex
	ttl       time.Duration
	states    map[string]time.Time
	now       func() time.Time
	byteCount int
}

func NewMemoryStateStore(cfg config.Config) *MemoryStateStore {
	ttl := cfg.Google.StateTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &MemoryStateStore{
		ttl:       ttl,
		states:    make(map[string]time.Time),
		now:       time.Now,
		byteCount: 32,
	}
}

func (store *MemoryStateStore) Generate() (string, error) {
	random := make([]byte, store.byteCount)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	state := base64.RawURLEncoding.EncodeToString(random)
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked()
	store.states[state] = store.now().UTC().Add(store.ttl)
	return state, nil
}

func (store *MemoryStateStore) Verify(state string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked()

	expiresAt, ok := store.states[state]
	if !ok {
		return false
	}
	delete(store.states, state)
	return store.now().UTC().Before(expiresAt)
}

func (store *MemoryStateStore) cleanupLocked() {
	now := store.now().UTC()
	for state, expiresAt := range store.states {
		if !now.Before(expiresAt) {
			delete(store.states, state)
		}
	}
}
