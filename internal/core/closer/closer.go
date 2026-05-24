package closer

import (
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type closeFn func() error

type namedCloser struct {
	name string
	fn   closeFn
}

// Manager is a simple LIFO closer with logging.
// Close executes registered closers in reverse registration order.
type Manager struct {
	mu      sync.Mutex
	closers []namedCloser
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) Add(name string, fn func() error) {
	if fn == nil {
		return
	}

	m.mu.Lock()
	m.closers = append(m.closers, namedCloser{name: name, fn: fn})
	m.mu.Unlock()
}

func (m *Manager) Close() error {
	m.mu.Lock()
	closers := make([]namedCloser, len(m.closers))
	copy(closers, m.closers)
	m.mu.Unlock()

	var err error
	for i := len(closers) - 1; i >= 0; i-- {
		c := closers[i]

		start := time.Now()
		log.Info().Str("resource", c.name).Msg("closing resource")
		if closeErr := c.fn(); closeErr != nil {
			log.Error().Str("resource", c.name).Err(closeErr).Dur("took", time.Since(start)).Msg("resource close failed")
			err = errors.Join(err, closeErr)
			continue
		}
		log.Info().Str("resource", c.name).Dur("took", time.Since(start)).Msg("resource closed")
	}

	return err
}
