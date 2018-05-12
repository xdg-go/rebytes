[![GoDoc](https://godoc.org/github.com/xdg/rebytes?status.svg)](https://godoc.org/github.com/xdg/rebytes)
[![Build Status](https://travis-ci.org/xdg/rebytes.svg?branch=master)](https://travis-ci.org/xdg/rebytes)

# Rebytes – Recycled byte slice pool and buffer for Golang

## Synopsis

```
    // Byte slice pool with 1k slice capacity, 100 pool capacity
    pool := rebytes.NewPool(1024, 100)
    byteSlice := pool.Get()
    pool.Put(byteSlice)

    // Like bytes.Buffer, but with dynamic memory managed by the pool
    buf, err := rebytes.NewBuffer(pool)
    buf.WriteString("Hello World")
```

## Status

This software is considered ALPHA quality.  Not recommended for production
use.

## Description

Package rebytes provides types that recycle bytes slices to reduce
allocation and garbage collection:

* `rebytes.Pool`: a `[]byte` pool that provides/recycles fixed capacity
  slices from a fixed-maximum-size pool.
* `rebytes.Buffer`: a `bytes.Buffer` analogue that dynamically gets/returns
  memory from a `rebytes.Pool`.

## Copyright and License

Copyright 2017 by David A. Golden. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License"). You may
obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
