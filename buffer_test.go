// Copyright 2017 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package rebytes_test

import (
	"io"
	"testing"

	"github.com/xdg/rebytes"
	"github.com/xdg/testy"
)

func TestNewBuffer(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	pool := rebytes.NewPool(5, 100)
	buf, err := rebytes.NewBuffer(pool)
	is.NotNil(buf)
	is.Nil(err)
}

func TestBufferWrite(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	// small slices size to ensure see multiple chunks
	pool := rebytes.NewPool(5, 100)
	buf, _ := rebytes.NewBuffer(pool)

	input := []byte("hel")
	n, err := buf.Write(input)
	is.Equal(n, len(input))
	is.Nil(err)
	is.Equal(buf.String(), "hel")
	is.Equal(buf.Chunks(), 1)

	input = []byte("lo world")
	n, err = buf.Write(input)
	is.Equal(n, len(input))
	is.Nil(err)
	is.Equal(buf.String(), "hello world")
	is.Equal(buf.Chunks(), 3)
}

func TestBufferWriteString(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	// small slices size to ensure see multiple chunks
	pool := rebytes.NewPool(5, 100)
	buf, _ := rebytes.NewBuffer(pool)

	input := "hello world"
	n, err := buf.WriteString(input)
	is.Equal(n, len(input))
	is.Nil(err)
	is.Equal(buf.String(), "hello world")
	is.Equal(buf.Chunks(), 3)
}

func TestBufferRead(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	pool := rebytes.NewPool(5, 100)
	buf, _ := rebytes.NewBuffer(pool)
	input := "hello world"
	n, err := buf.WriteString(input)
	is.Equal(buf.Chunks(), 3)

	got := make([]byte, 0)
	n, err = buf.Read(got)
	is.Equal(n, 0)
	is.Nil(err)

	got = make([]byte, 3)
	n, err = buf.Read(got)
	is.Equal(n, 3)
	is.Nil(err)
	is.Equal(string(got), "hel")
	is.Equal(buf.Chunks(), 3)

	n, err = buf.Read(got)
	is.Equal(n, 3)
	is.Nil(err)
	is.Equal(string(got), "lo ")
	is.Equal(buf.Chunks(), 3)

	got = make([]byte, 20)
	n, err = buf.Read(got)
	got = got[0:n]
	is.Equal(n, 5)
	is.NotNil(err)
	is.True(err == io.EOF)
	is.Equal(string(got), "world")
	is.Equal(buf.Chunks(), 3)
}

func TestBufferError(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	buf, err := rebytes.NewBuffer(nil)
	is.Nil(buf)
	is.NotNil(err)
	is.Equal(err.Error(), "can't pass nil to rebytes.NewBuffer()")
}

func initTestBuffer(is *testy.T) *rebytes.Buffer {
	pool := rebytes.NewPool(5, 100)
	buf, _ := rebytes.NewBuffer(pool)

	// load buffer
	_, err := buf.WriteString("hello world")
	is.Nil(err)
	is.Equal(buf.Chunks(), 3)

	return buf
}

func TestBufferFree(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	buf := initTestBuffer(is)

	// free buffer
	buf.Free()
	is.Equal(buf.Chunks(), 0)

	// read after free is error
	got := make([]byte, 3)
	_, err := buf.Read(got)
	is.Equal(err.Error(), "can't read from buffer after Free()")

	// write after free is error
	_, err = buf.WriteString("hello world")
	is.Equal(err.Error(), "can't write to buffer after Free()")
}

func TestBufferReadAt(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	buf := initTestBuffer(is)

	s := make([]byte, 3)
	n, err := buf.ReadAt(s, 0)
	is.Equal(string(s), "hel")
	is.Equal(n, 3)
	is.Nil(err)

	got := make([]byte, 20)
	n, err = buf.ReadAt(got, 6)
	got = got[0:n]
	is.Equal(n, 5)
	is.NotNil(err)
	is.True(err == io.EOF)
	is.Equal(string(got), "world")
}
