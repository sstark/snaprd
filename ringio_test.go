/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
	"bytes"
	"reflect"
	"testing"
)

type rioTestPair struct {
	params [2]int
	in     [][]byte
	out    []byte
}

func TestRingIO(t *testing.T) {
	tests := []rioTestPair{
		{
			[2]int{3, 12},
			[][]byte{
				[]byte("a string\n"),
				[]byte("another string\n"),
				[]byte("something\n"),
				[]byte("else\n"),
			},
			[]byte("another stri\nsomething\nelse\n"),
		},
		{
			[2]int{2, 10},
			[][]byte{
				[]byte("a string\n"),
				[]byte("another string\n"),
				[]byte("something\n"),
				[]byte("else\n"),
			},
			[]byte("something\nelse\n"),
		},
		{
			[2]int{2, 4},
			[][]byte{
				[]byte("a string"),
				[]byte("test1"),
				[]byte("test2"),
			},
			[]byte("test\ntest\n"),
		},
		{
			[2]int{100, 100},
			[][]byte{
				[]byte("a string"),
				[]byte("test1"),
				[]byte("test2"),
			},
			[]byte("a stringtest1test2"),
		},
	}
	var buf bytes.Buffer
	for _, tp := range tests {
		rio := newRingIO(&buf, tp.params[0], tp.params[1])
		for _, l := range tp.in {
			rio.Write(l)
		}
		got := rio.GetAsText()
		wanted := tp.out
		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("wanted:\n>>>\n%s\n<<<\ngot:\n>>>\n%s\n<<<", wanted, got)
		}
	}
}
