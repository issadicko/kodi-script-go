// Package cache provides LRU caching for parsed AST.
package cache

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/issadicko/kodi-script-go/ast"
)

// ASTCache is an LRU cache for parsed AST programs.
type ASTCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type cacheEntry struct {
	key     string
	program *ast.Program
}

// NewASTCache creates a new AST cache with the given capacity.
func NewASTCache(capacity int) *ASTCache {
	return &ASTCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// hash generates a hash key for the source code.
func hash(source string) string {
	h := sha256.Sum256([]byte(source))
	return hex.EncodeToString(h[:8]) // Use first 8 bytes for shorter key
}

// Get retrieves a cached AST program.
func (c *ASTCache) Get(source string) (*ast.Program, bool) {
	key := hash(source)

	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.order.MoveToFront(elem)

	return elem.Value.(*cacheEntry).program, true
}

// Set stores an AST program in the cache.
func (c *ASTCache) Set(source string, program *ast.Program) {
	key := hash(source)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheEntry).program = program
		return
	}

	// Evict if at capacity
	if c.order.Len() >= c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}

	// Add new entry
	entry := &cacheEntry{key: key, program: program}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

// Clear removes all entries from the cache.
func (c *ASTCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order.Init()
}

// Len returns the number of cached entries.
func (c *ASTCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// DefaultCache is a global AST cache with default capacity.
var DefaultCache = NewASTCache(1000)
