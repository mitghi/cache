/* MIT License
* 
* Copyright (c) 2018 Mike Taghavi <mitghi[at]gmail.com>
* 
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
* 
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
*/

package cache

import (
	"container/list"
	"sync"
)

// Ensure interface (protocol) conformance
var (
	_ CacheInterface = (*LRU)(nil)
)

// Defaults
const (
	defaultCAPACITY = 16
)

// LRU implements Least Recently Used
// caching policy.
type LRU struct {
	// size: 64 bytes
	mu              *sync.RWMutex                 // 8 bytes
	items           *list.List                    // 8 bytes
	lookup          map[interface{}]*list.Element // 8 bytes
	capacity, count int                           // 8 bytes
	_               [3]uint64                     // 24 bytes
}

// LRUItem is the container for
// individual cache enteries.
type LRUItem struct {
	// size: 64 bytes
	Key   interface{} // 16 bytes
	Value interface{} // 16 bytes
	Count int         // 8 bytes
	_     [3]uint64   // 24 bytes
}

// - MARK: Alloc/Init section.

// NewLRU allocates and initializes a new
// `LRU` struct and returns a pointer to it.
// Note, when `capacity <= 0` holds true,
// capacity is set to `defaultCAPACITY` (
// by default 16 ).
func NewLRU(capacity int) (lru *LRU) {
	lru = &LRU{
		mu:       &sync.RWMutex{},
		items:    list.New(),
		lookup:   make(map[interface{}]*list.Element),
		capacity: capacity - 1,
		count:    0,
	}
	// ensure validity of capacity
	if lru.capacity <= 0 {
		lru.capacity = defaultCAPACITY
	}
	return lru
}

// - MARK: LRU section.

// Set writes k/v pair in the cache and evicts
// old enteries when needed. It sets `isNew` to
// to `true` when the given k/v pair are allocated
// ( i.e. wasn't in cache ) and an error to indicate
// failures.
func (lru *LRU) Set(key interface{}, value interface{}) (isNew bool, err error) {
	lru.mu.Lock()
	isNew, err = lru.set(key, value)
	lru.mu.Unlock()
	return isNew, err
}

// Get fetches `key` from cache and return its value
// when available along with an error in case of
// failure.
func (lru *LRU) Get(key interface{}) (value interface{}, err error) {
	var (
		item *LRUItem
	)
	value = nil
	lru.mu.Lock()
	// only return value to prevent
	// data race
	item, err = lru.get(key)
	if err == nil && item != nil {
		value = item.Value
	}
	lru.mu.Unlock()
	return value, err
}

// Read only reads the given item with `key` without
// incrementing cache counter or triggering eviction
// policies. When no item with given `key` exists,
// it returns `nil`.
func (lru *LRU) Read(key interface{}) (value interface{}) {
	var (
		item *LRUItem
	)
	lru.mu.Lock()
	item = lru.read(key)
	if item != nil {
		value = item.Value
	}
	lru.mu.Unlock()
	return value
}

// Remove removes the given item with `key` from cache
// and returns `true` when succesfull.
func (lru *LRU) Remove(key interface{}) (ok bool) {
	lru.mu.Lock()
	ok = lru.remove(key)
	lru.mu.Unlock()
	return ok
}

// Purge removes all enteries and restarts the cache.
func (lru *LRU) Purge() {
	lru.mu.Lock()
	lru.reset()
	lru.mu.Unlock()
}

// Len returns number of items in cache.
func (lru *LRU) Len() (l int) {
	lru.mu.Lock()
	l = lru.items.Len()
	lru.mu.Unlock()
	return l
}

// set writes k/v pair in the cache and triggers
// eviction policies when neccessary. Note, this
// routine is not protected against concurrent
// accesses; therefore not publicly exposed.
func (lru *LRU) set(key interface{}, value interface{}) (isNew bool, err error) {
	// increment global LRU counter
	lru.count++
	var (
		cnt  int = lru.items.Len()
		item *LRUItem
		elem *list.Element
		ok   bool
	)
	elem, ok = lru.lookup[key]
	if !ok {
		if cnt > lru.capacity {
			lru.evict()
		}
		isNew = true
		item = &LRUItem{Count: lru.count, Key: key, Value: value}
		elem = lru.items.PushFront(item)
		lru.lookup[key] = elem
		goto OK
	}
	item, ok = elem.Value.(*LRUItem)
	if !ok {
		err = ELRUINVALTYPE
		goto ERROR
	}
	if cnt > lru.capacity {
		lru.evict()
	}
	item.Count += 1
	item.Value = value
	lru.items.MoveToFront(elem)

OK:
	return isNew, nil
ERROR:
	return false, err
}

// get fetches the item associated to given `key`. Note,
// this routine is not protected against concurrent
// accesses; therefore not publicly exposed.
func (lru *LRU) get(key interface{}) (value *LRUItem, err error) {
	lru.count++
	var (
		item *LRUItem
		elem *list.Element
		ok   bool
	)
	elem, ok = lru.lookup[key]
	if !ok {
		goto ERROR
	}
	item = elem.Value.(*LRUItem)
	item.Count++
	lru.items.MoveToFront(elem)

	return item, nil
ERROR:
	return nil, err
}

// readEntery returns the element associated to `key`
// when available without triggering eviction policies
// and incrementing cache counters. Note, this routine
// is not protected against concurrent accesses; therefore
// not publicly exposed.
func (lru *LRU) readEntery(key interface{}) *list.Element {
	var (
		elem *list.Element
		ok   bool
	)
	elem, ok = lru.lookup[key]
	if !ok {
		return nil
	}
	return elem
}

// read returns the `*LRUItem` associated to `key`
// when avaialble without triggering eviction policies
// and incrementing cache counters. Note, this routine
// is not protected against concurrent accesses. therefore
// not publicly exposed.
func (lru *LRU) read(key interface{}) *LRUItem {
	var (
		elem *list.Element
	)
	elem = lru.lookup[key]
	if elem == nil {
		return nil
	}
	return elem.Value.(*LRUItem)
}

// reset purges all cache enteries and restarts
// the internal state to a freshly initialized
// instance. Note, this routine is not protected
// against concurrent accesses; therefore not
// publicly exposed.
func (lru *LRU) reset() {
	lru.items = lru.items.Init()
	lru.count = 0
	for k, _ := range lru.lookup {
		delete(lru.lookup, k)
	}
}

// remove removes the entery associated to the
// given `key` without invoking caching policies
// or incrementing counters. It returns true
// when successfull. Note, this routine is not
// protected against concurrent accesses; therefore
// not publicly exposed.
func (lru *LRU) remove(key interface{}) bool {
	var (
		item *list.Element = lru.readEntery(key)
	)
	if item == nil {
		return false
	}
	// TODO:
	// . remove references from
	//   node before returning?
	_ = lru.items.Remove(item)
	return true
}

// evict is the policy function. It removes
// oldest entery ( i.e. pops an item from back
// of the list ) and removes its references.
// Note, this routine is not protected against
// concurrent accesses; therefore not publicly
// exposed.
func (lru *LRU) evict() {
	var (
		item *LRUItem = lru.popBack()
	)
	delete(lru.lookup, item.Key)
	// remove references to help GC
	item.Key = nil
	item.Value = nil
	item = nil
}

// popBack removes tail item. Note, this routine
// is not protected agaisnt concurrent accesses;
// therefore not publicy exopsed.
func (lru *LRU) popBack() (elem *LRUItem) {
	return lru.items.Remove(lru.items.Back()).(*LRUItem)
}

// popFront removes head item. Note, this routine
// is not protected against concurrent accesses;
// therefore not publicly exposed.
func (lru *LRU) popFront() (elem *LRUItem) {
	return lru.items.Remove(lru.items.Front()).(*LRUItem)
}

// dump is a receiver that dumps all items
// from the container. Note, this routine
// is not protected against cocurrent accesses;
// therefore not publicly exposed.
func (lru *LRU) dump() (items []*LRUItem, err error) {
	var (
		cnt   int = lru.items.Len()
		value *LRUItem
	)
	items = make([]*LRUItem, 0, cnt)
	for i := 0; i < cnt; i++ {
		value = lru.items.Remove(lru.items.Front()).(*LRUItem)
		if value == nil {
			err = ELRUFATAL
			goto ERROR
		}
		items = append(items, value)
	}
	return items, nil
ERROR:
	return nil, err
}

// - MARK: LRUItem section.

// K conforms to `CacheItemInterface` and returns
// associated key.
func (lrui *LRUItem) K() interface{} {
	return lrui.Key
}

// V conforms to `CacheItemInterface` and returns
// associated value.
func (lrui *LRUItem) V() interface{} {
	return lrui.Value
}

// C conforms to `CacheItemInterface` and returns
// associated count.
func (lrui *LRUItem) C() interface{} {
	return lrui.Count
}
