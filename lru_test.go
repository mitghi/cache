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
	"fmt"
	"testing"

	"github.com/mitghi/x/structs"
)

func init() {
	fmt.Println(structs.CompileStructInfo(LRU{}))
	fmt.Println(structs.CompileStructInfo(LRUItem{}))
}

func hasPositiveLRUS(s []*LRU, mask byte) bool {
	var (
		bit byte
	)
	for i := uint(0); i < uint(len(s)); i++ {
		bit = (mask & (0x1 << i)) >> i
		if bit == 0x1 {
			if s[i].capacity < 0 {
				return false
			}
		}
	}
	return true
}

func TestLRUCapacity(t *testing.T) {
	const length int = 4
	var (
		mask       byte = 0x09 // 0b1001
		lrus       []*LRU
		capacities []int
	)
	capacities = []int{
		0, 8, 16, -1,
	}
	for _, v := range capacities {
		lrus = append(lrus, NewLRU(v))
	}
	if !hasPositiveLRUS(lrus, mask) {
		t.Fatal("assertion failed, inconsistent state. expected true.")
	}
}

func TestLRU(t *testing.T) {
	const (
		defCAPACITY int    = 8
		iters       int    = 10
		usrfmt      string = "user_%d"
	)
	var (
		lru   *LRU = NewLRU(defCAPACITY)
		sfmt  string
		value interface{}
		err   error
	)
	for i := 0; i < iters; i++ {
		var (
			isNew bool
			err   error
		)
		sfmt = fmt.Sprintf(usrfmt, i)
		isNew, err = lru.Set(sfmt, i)
		if !isNew || err != nil {
			t.Fatal("inconsistent state, expected equal.")
		}
	}
	// assert container size
	if l := lru.Len(); l != defCAPACITY {
		t.Fatalf("assertion failed; inconsistent state, expected equal with value(%d) - got value(%d).", defCAPACITY, l)
	}
	value, err = lru.Get("user_8")
	if value == nil || err != nil {
		t.Fatal("inconsistent state, expected equal.", value, err)
	}
	value = lru.Read("user_8")
	if value == nil {
		t.Fatal("inconsistent state, expected unequal.")
	}
	if lru.items.Front().Value.(*LRUItem).Key.(string) == "user_0" {
		t.Fatal("assertion failed, expected unequal.")
	}
	if lru.Len() != 8 {
		t.Fatal("assertion failed, inconsistent state. expected equal.")
	}
	// remove a value from cache
	if !lru.Remove("user_8") {
		t.Fatal("assertion failed, inconsistent state. expected equal.")
	}
	if lru.Len() != 7 {
		t.Fatal("assertion failed, inconsistent state. expected equal.")
	}
	lru.Purge()
	if lru.count != 0 || lru.items.Len() != 0 || len(lru.lookup) != 0 {
		t.Fatal("assertion afiled, inconsistent state, expected equal.", lru.count, lru.items.Len(), lru.lookup)
	}
}
