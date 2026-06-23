// Package cache provides LRU caching for parsed AST.
package cache

import (
	"container/list"
	"sync"

	"github.com/issadicko/kodi-script-go/ast"
)

// ASTCache is an LRU cache for parsed AST programs, keyed directly by source.
// Using the source string as the key avoids hashing on every lookup and the
// hash-collision handling that a digest key would require (Go already hashes
// the string for the map internally, and that hash is cached on the string).
type ASTCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type cacheEntry struct {
	source  string
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

// Get retrieves a cached AST program.
func (c *ASTCache) Get(source string) (*ast.Program, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[source]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(elem)
	return elem.Value.(*cacheEntry).program, true
}

// Set stores an AST program in the cache.
func (c *ASTCache) Set(source string, program *ast.Program) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[source]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheEntry).program = program
		return
	}

	// Evict the least-recently-used entry if at capacity.
	if c.order.Len() >= c.capacity {
		if oldest := c.order.Back(); oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).source)
		}
	}

	c.items[source] = c.order.PushFront(&cacheEntry{source: source, program: program})
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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// DefaultCache is a global AST cache with default capacity.
var DefaultCache = NewASTCache(1000)
