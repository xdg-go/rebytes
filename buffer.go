// Copyright 2017 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

// Package rebytes provides types that recycle bytes slices to reduce
// allocation and garbage collection.
package rebytes

import (
	"bytes"
	"errors"
	"io"
)

// Buffer is a dynamic buffer of bytes satisfying the Reader and Writer
// interfaces.  Memory is managed by a rebytes.Pool: new memory is allocated
// from the pool as needed for writign and returned to the pool after reading.
type Buffer struct {
	pool     *Pool
	offset   int // where to read from
	chunks   [][]byte
	chunkCap int // capacity of a chunk
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data
// in a buffer.
var ErrTooLarge = errors.New("rebytes.Buffer: too large")

// NewBuffer creates a new buffer from a rebytes.Pool.  It only errors if
// passed nil.
func NewBuffer(pool *Pool) (*Buffer, error) {
	if pool == nil {
		return nil, errors.New("can't pass nil to rebytes.NewBuffer()")
	}
	b := &Buffer{pool: pool}
	b.chunks = make([][]byte, 1)
	b.chunks[0] = pool.Get()
	b.chunkCap = cap(b.chunks[0])
	return b, nil
}

// findReadLocation returns a pointer to the chunk corresponding to
// the current offset and a chunk-relative offset to read from
func (b *Buffer) findReadLocation(offset int) (s *[]byte, i int) {
	s = &b.chunks[offset/b.chunkCap]
	i = offset % b.chunkCap
	return
}

// findWritableChunk returns a pointer to a writeable chunk, adding one if
// needed.
func (b *Buffer) findWritableChunk() (w *[]byte) {
	w = &b.chunks[len(b.chunks)-1]
	l, c := len(*w), cap(*w)
	if c-l == 0 {
		s := b.pool.Get()
		b.chunks = append(b.chunks, s)
		return &b.chunks[len(b.chunks)-1]
	}
	return w
}

// Chunks returns the number of []byte chunks used by the buffer.
func (b *Buffer) Chunks() int {
	return len(b.chunks)
}

// Free recycles byte slices back into the associated pool.  The
// buffer is unusable afterwards.
func (b *Buffer) Free() {
	for _, s := range b.chunks {
		b.pool.Put(s)
	}
	b.chunks = b.chunks[0:0]
}

// String joins unread []byte chunks into single string
func (b *Buffer) String() string {
	return string(bytes.Join(b.chunks, []byte("")))
}

// Read reads the next len(p) unread bytes into p.  It returns the number of
// bytes read (0 <= n <= len(p)) and any error encountered.  When there is
// too-little data to fill p, the error will be io.EOF.
func (b *Buffer) Read(p []byte) (n int, err error) {
	n, err = b.readFromOffset(p, b.offset)
	b.offset += n
	return
}

// ReadAt reads len(p) bytes into p starting at offset off in the buffer. It
// returns the number of bytes read (0 <= n <= len(p)) and any error
// encountered.  When there is too-little data to fill p, the error will be
// io.EOF.
func (b *Buffer) ReadAt(p []byte, off int64) (n int, err error) {
	return b.readFromOffset(p, int(off))
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) Write(p []byte) (n int, err error) {
	if len(b.chunks) == 0 {
		return 0, errors.New("can't write to buffer after Free()")
	}

	for len(p) > 0 {
		w := b.findWritableChunk()
		n += moveBytes(w, &p)
	}

	return n, nil
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) WriteString(s string) (n int, err error) {
	return b.Write([]byte(s))
}

// ReaderFrom is the interface that wraps the ReadFrom method.
//
// ReadFrom reads data from r until EOF or error.
// The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
//
// The Copy function uses ReaderFrom if available.
// type ReaderFrom interface {
// 	ReadFrom(r Reader) (n int64, err error)
// }

// WriterTo is the interface that wraps the WriteTo method.
//
// WriteTo writes data to w until there's no more data to write or
// when an error occurs. The return value n is the number of bytes
// written. Any error encountered during the write is also returned.
//
// The Copy function uses WriterTo if available.
// type WriterTo interface {
// 	WriteTo(w Writer) (n int64, err error)
// }

// ByteScanner is the interface that adds the UnreadByte method to the
// basic ReadByte method.
//
// UnreadByte causes the next call to ReadByte to return the same byte
// as the previous call to ReadByte.
// It may be an error to call UnreadByte twice without an intervening
// call to ReadByte.
// type ByteScanner interface {
// 	ByteReader
// 	UnreadByte() error
// }

// ByteWriter is the interface that wraps the WriteByte method.
// type ByteWriter interface {
// 	WriteByte(c byte) error
// }

// func (b *Buffer) Len() int { return len(b.buf) - b.off }
// func (b *Buffer) Truncate(n int) {
// func (b *Buffer) Reset() { b.Truncate(0) }

// func (b *Buffer) Next(n int) []byte {

// func (b *Buffer) ReadBytes(delim byte) (line []byte, err error) {
// func (b *Buffer) ReadString(delim byte) (line string, err error) {

func (b *Buffer) readFromOffset(p []byte, off int) (n int, err error) {
	if len(b.chunks) == 0 {
		return 0, errors.New("can't read from buffer after Free()")
	}

	for n < len(p) {
		s, i := b.findReadLocation(off)

		if i == len(*s) {
			return n, io.EOF
		}

		// copy and then update offset and progress counter
		x := copy(p[n:], (*s)[i:])
		off += x
		n += x
	}

	return n, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// moveBytes destructively copies from source to destination; the
// destination is grown to fit; the source is resliced to omit
// the part copied
func moveBytes(dst, src *[]byte) (n int) {
	// shortcut if nothing to move
	if len(*src) == 0 {
		return 0
	}

	// compute unused capacity remaining
	l, c := len(*dst), cap(*dst)
	r := c - l

	// panic if no remaining capacity to move to
	if r == 0 {
		panic(ErrTooLarge)
	}

	// length to move is smaller of src length or remaining capacity
	r = min(len(*src), r)

	// reslice dst to get space to copy
	*dst = (*dst)[:l+r]

	// copy target number of bytes from src to dst
	copy((*dst)[l:], (*src)[:r])

	// reslice source past bytes copied
	*src = (*src)[r:]

	return r
}
