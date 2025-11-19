package localqueue

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

const (
	bucketName = "queue"
	// StatusPending indicates an item is waiting to be sent
	StatusPending = "pending"
	// StatusSent indicates an item has been successfully sent
	StatusSent = "sent"
	// DefaultMaxItems is the default maximum number of items in the queue
	DefaultMaxItems = 10000
	// DefaultItemTTL is the default time-to-live for items (7 days)
	DefaultItemTTL = 7 * 24 * time.Hour
	// DefaultCleanupInterval is the default interval for cleanup operations
	DefaultCleanupInterval = 1 * time.Hour
)

var (
	// ErrQueueFull is returned when the queue has reached its maximum capacity
	ErrQueueFull = errors.New("local queue full")
	// ErrItemNotFound is returned when an item with the given ID is not found
	ErrItemNotFound = errors.New("item not found")
	// ErrInvalidItem is returned when an item has invalid data
	ErrInvalidItem = errors.New("invalid item")
)

// Item represents an entry in the local queue.
type Item struct {
	ID        string `json:"id"`
	Payload   []byte `json:"payload"`
	Status    string `json:"status"` // pending/sent
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at,omitempty"` // Unix timestamp, 0 means no expiration
}

// LocalQueue is a persistent queue backed by bbolt for storing messages
// when network connectivity is unavailable.
type LocalQueue struct {
	db              *bbolt.DB
	maxItems        int
	itemTTL         time.Duration
	cleanupInterval time.Duration
}

// NewLocalQueue opens or creates a bbolt DB at path.
// If maxItems <= 0, DefaultMaxItems is used.
// If itemTTL <= 0, DefaultItemTTL is used.
func NewLocalQueue(path string, maxItems int, itemTTL time.Duration) (*LocalQueue, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		return e
	}); err != nil {
		db.Close()
		return nil, err
	}
	if maxItems <= 0 {
		maxItems = DefaultMaxItems
	}
	if itemTTL <= 0 {
		itemTTL = DefaultItemTTL
	}
	return &LocalQueue{
		db:              db,
		maxItems:        maxItems,
		itemTTL:         itemTTL,
		cleanupInterval: DefaultCleanupInterval,
	}, nil
}

// Close closes the underlying DB.
func (q *LocalQueue) Close() error {
	return q.db.Close()
}

// Enqueue stores a payload as a pending item and returns its id.
// Returns ErrQueueFull if the queue has reached its maximum capacity.
func (q *LocalQueue) Enqueue(payload []byte) (string, error) {
	if payload == nil {
		return "", errors.New("payload cannot be nil")
	}
	count, err := q.Count()
	if err != nil {
		return "", err
	}
	if count >= q.maxItems {
		return "", ErrQueueFull
	}
	id := uuid.New().String()
	now := time.Now()
	expiresAt := int64(0)
	if q.itemTTL > 0 {
		expiresAt = now.Add(q.itemTTL).Unix()
	}
	item := Item{
		ID:        id,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: now.Unix(),
		ExpiresAt: expiresAt,
	}
	data, err := json.Marshal(item)
	if err != nil {
		return "", err
	}
	if err := q.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put([]byte(id), data)
	}); err != nil {
		return "", err
	}
	return id, nil
}

// DequeuePending returns up to limit pending items that have not expired.
// If limit <= 0, a default limit of 10 is used.
// Expired items are automatically filtered out.
func (q *LocalQueue) DequeuePending(limit int) ([]Item, error) {
	items := make([]Item, 0, limit)
	if limit <= 0 {
		limit = 10
	}
	now := time.Now().Unix()
	err := q.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil && len(items) < limit; k, v = c.Next() {
			var it Item
			if err := json.Unmarshal(v, &it); err != nil {
				continue
			}
			// Skip expired items
			if it.ExpiresAt > 0 && it.ExpiresAt < now {
				continue
			}
			if it.Status == StatusPending {
				items = append(items, it)
			}
		}
		return nil
	})
	return items, err
}

// MarkSent marks item as sent (so it won't be retried).
// Returns ErrItemNotFound if the item with the given ID does not exist.
func (q *LocalQueue) MarkSent(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	return q.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(id))
		if v == nil {
			return ErrItemNotFound
		}
		var it Item
		if err := json.Unmarshal(v, &it); err != nil {
			return ErrInvalidItem
		}
		it.Status = StatusSent
		data, err := json.Marshal(it)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}

// Delete removes an item from the queue by ID.
// Returns ErrItemNotFound if the item does not exist.
func (q *LocalQueue) Delete(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	return q.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b.Get([]byte(id)) == nil {
			return ErrItemNotFound
		}
		return b.Delete([]byte(id))
	})
}

// Get retrieves an item by ID.
// Returns ErrItemNotFound if the item does not exist.
func (q *LocalQueue) Get(id string) (*Item, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	var item *Item
	err := q.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(id))
		if v == nil {
			return ErrItemNotFound
		}
		var it Item
		if err := json.Unmarshal(v, &it); err != nil {
			return ErrInvalidItem
		}
		item = &it
		return nil
	})
	return item, err
}

// Count returns number of pending items that have not expired.
func (q *LocalQueue) Count() (int, error) {
	cnt := 0
	now := time.Now().Unix()
	err := q.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var it Item
			if err := json.Unmarshal(v, &it); err != nil {
				continue
			}
			// Skip expired items
			if it.ExpiresAt > 0 && it.ExpiresAt < now {
				continue
			}
			if it.Status == StatusPending {
				cnt++
			}
		}
		return nil
	})
	return cnt, err
}

// CountAll returns the total number of items (both pending and sent) that have not expired.
func (q *LocalQueue) CountAll() (int, error) {
	cnt := 0
	now := time.Now().Unix()
	err := q.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var it Item
			if err := json.Unmarshal(v, &it); err != nil {
				continue
			}
			// Skip expired items
			if it.ExpiresAt > 0 && it.ExpiresAt < now {
				continue
			}
			cnt++
		}
		return nil
	})
	return cnt, err
}

// Cleanup removes expired items and optionally removes sent items older than a specified duration.
// If removeSentOlderThan > 0, sent items older than this duration will also be removed.
// Returns the number of items removed.
func (q *LocalQueue) Cleanup(removeSentOlderThan time.Duration) (int, error) {
	removed := 0
	
	err := q.db.Update(func(tx *bbolt.Tx) error {
		// Calculate time inside transaction for consistency
		now := time.Now()
		nowUnix := now.Unix()
		cutoffTime := now.Add(-removeSentOlderThan).Unix()
		
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		keysToDelete := make([][]byte, 0)
		
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var it Item
			if err := json.Unmarshal(v, &it); err != nil {
				// Invalid item, mark for deletion - copy key to avoid cursor issues
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				keysToDelete = append(keysToDelete, keyCopy)
				continue
			}
			
			// Remove expired items
			if it.ExpiresAt > 0 && it.ExpiresAt < nowUnix {
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				keysToDelete = append(keysToDelete, keyCopy)
				continue
			}
			
			// Remove old sent items if requested
			if removeSentOlderThan > 0 && it.Status == StatusSent && it.CreatedAt < cutoffTime {
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				keysToDelete = append(keysToDelete, keyCopy)
				continue
			}
		}
		
		// Delete all marked keys
		for _, key := range keysToDelete {
			if err := b.Delete(key); err != nil {
				return err
			}
			removed++
		}
		
		return nil
	})
	
	return removed, err
}

// SetCleanupInterval sets the interval for periodic cleanup operations.
func (q *LocalQueue) SetCleanupInterval(interval time.Duration) {
	if interval > 0 {
		q.cleanupInterval = interval
	}
}

// GetCleanupInterval returns the current cleanup interval.
func (q *LocalQueue) GetCleanupInterval() time.Duration {
	return q.cleanupInterval
}
