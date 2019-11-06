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

// Package cache provides common caching facilities.
package cache

import "errors"

// Error messages
var (
	ELRUINVALTYPE error = errors.New("cache(lru): invalid item type.")
	ELRUFATAL     error = errors.New("cache(lru): fatal state.")
)

// CacheInterface is protocol definition that
// must be conformed when implementing cache
// subsubsystem.
type CacheInterface interface {
	Set(interface{}, interface{}) (bool, error)
	Get(interface{}) (interface{}, error)
	Read(interface{}) interface{}
	Purge()
	Len() int
}

// CacheItemInterface is protocol definition
// for indiviudal items in cache lines that
// must be conformed.
type CacheItemInterface interface {
	K() interface{}
	V() interface{}
	C() int
}
