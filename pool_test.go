// Copyright 2017 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package rebytes_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/xdg/rebytes"
	"github.com/xdg/testy"
)

const (
	bufSize  = 256
	poolSize = 10
)

// utility function for use in comparing if slices are the same
func getSliceHeader(b *[]byte) *reflect.SliceHeader {
	return (*reflect.SliceHeader)(unsafe.Pointer(b))
}

func TestNewPool(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	pool := rebytes.NewPool(bufSize, poolSize)
	is.NotNil(pool)
	is.Equal(pool.Size(), 0)
}

func TestPoolGet(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	pool := rebytes.NewPool(bufSize, poolSize)

	for _ = range []int{1, 2, 3} {
		b := pool.Get()
		is.NotNil(b)
		is.Equal(pool.Size(), 0)
		is.Equal(len(b), 0)
		is.Equal(cap(b), bufSize)
	}

	b1 := pool.Get()
	b2 := pool.Get()
	is.Unequal(getSliceHeader(&b1), getSliceHeader(&b2))
}

func TestPoolPut(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	pool := rebytes.NewPool(bufSize, poolSize)

	// Check that Get-Put-Get gives back original underlying []byte data
	// but with slice length reset to zero
	b1 := pool.Get()
	b1 = append(b1, []byte("abc")...)
	is.Equal(len(b1), 3)
	pool.Put(b1)
	is.Equal(pool.Size(), 1)

	b2 := pool.Get()
	is.Equal(getSliceHeader(&b1).Data, getSliceHeader(&b2).Data)
	is.Equal(len(b2), 0)
	is.Equal(cap(b2), bufSize)
	is.Equal(pool.Size(), 0)

	// Fill up the pool
	slices := make([][]byte, poolSize+1)
	for i := 0; i < poolSize+1; i++ {
		slices[i] = pool.Get()
	}

	for i := 0; i < poolSize; i++ {
		pool.Put(slices[i])
		is.Equal(pool.Size(), i+1)
	}

	// Check that extra puts are dropped
	pool.Put(slices[poolSize])
	is.Equal(pool.Size(), poolSize)
	b3 := pool.Get()
	is.Equal(getSliceHeader(&b3), getSliceHeader(&slices[poolSize-1]))
}

func TestPoolError(t *testing.T) {
	is := testy.New(t)
	defer func() { t.Logf(is.Done()) }()

	pool := rebytes.NewPool(bufSize, poolSize)
	var err error

	// Put nil
	err = pool.Put(nil)
	is.Equal(err.Error(), "nil value not allowed for rebytes.Pool.Put()")
	is.Equal(pool.Size(), 0)

	// Put wrong capacity slice back in
	cases := map[string][]byte{
		"too large": make([]byte, 0, bufSize+1),
		"too small": make([]byte, 0, bufSize-1),
	}
	for k, v := range cases {
		is := is.Label(k)
		err = pool.Put(v)
		is.NotNil(err)
		is.Equal(err.Error(), "wrong capacity byte slice for rebytes.Pool.Put()")
		is.Equal(pool.Size(), 0)
	}
}
