/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
	"bytes"
	"io"
	"sync"
)

type RingIO struct {
	out     io.Writer // the io we are proxying
	maxLen  int       // max number of lines
	maxElem int       // max length of a line
	mu      *sync.Mutex
	buf     map[int][]byte // a map holding the lines
	p       int            // points to the current item in the map
}

// newRingIO instantiates a new RingIO list, which satisfies the io.Writer
// out is an io.Writer that will write the output to the final destination.
// maxLen will be the maximum number of elements kept in the ring buffer. If
// this number is reached, for each Write() the first element will be removed
// before the new element is added.
// maxElem is the maximum size in bytes of an individual element of the list.
func newRingIO(out io.Writer, maxLen int, maxElem int) *RingIO {
	return &RingIO{
		out:     out,
		maxLen:  maxLen,
		maxElem: maxElem,
		mu:      new(sync.Mutex),
		buf:     make(map[int][]byte),
		p:       0,
	}
}

func (r *RingIO) Write(s []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var e []byte
	// we need to copy the slice because the caller may be reusing it
	c := make([]byte, len(s))
	copy(c, s)
	// write to the io.Writer we are proxying
	r.out.Write(c)
	ls := len(c)
	// if needed, truncate the new entry to maxElem bytes and append a newline
	if ls > r.maxElem {
		e = append(c[0:r.maxElem], byte('\n'))
	} else {
		e = c
	}
	r.buf[r.p] = e
	// reset the pointer if maxLen is reached
	if r.p < r.maxLen-1 {
		r.p += 1
	} else {
		r.p = 0
	}
	return len(c), nil
}

// GetAll returns all elements of the ring buffer as a slice of byte slices
func (r *RingIO) GetAll() [][]byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ret [][]byte
	// return buf, but starting from where the pointer currently points to
	for i := r.p; i < r.maxLen; i += 1 {
		ret = append(ret, r.buf[i])
	}
	for i := 0; i < r.p; i += 1 {
		ret = append(ret, r.buf[i])
	}
	return ret
}

// GetAsText concatenates all buffered lines into one byte slice
func (r *RingIO) GetAsText() []byte {
	var b bytes.Buffer
	for _, l := range r.GetAll() {
		b.Write(l)
	}
	return b.Bytes()
}
