package worker

import (
	"sync"

	"github.com/google/uuid"
)

type TopicLockManager struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

var globalTopicLocks = &TopicLockManager{
	locks: make(map[string]*sync.Mutex),
}

// getLock returns the mutex for a specific topic, creating one if it doesn't exist
func (m *TopicLockManager) getLock(topicID string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[topicID]
	if !exists {
		lock = &sync.Mutex{}
		m.locks[topicID] = lock
	}
	return lock
}

// LockTopic acquires the lock for a specific topic_id and returns an unlock function
func LockTopic(topicID uuid.UUID) func() {
	lock := globalTopicLocks.getLock(topicID.String())
	lock.Lock()
	return func() {
		lock.Unlock()
	}
}
