# Rebytes – Recycled byte slice pool and buffer for Golang

## Synopsis

```
    // Byte slice pool with 1k slice capacity, 100 pool capacity
    pool := rebytes.NewPool(1024, 100)
    byteSlice := pool.Got()
    pool.Put(byteSlice)

    // Like bytes.Buffer, but with dynamic memory managed by the pool
    buf, err := rebytes.NewBuffer(pool)
    buf.WriteString("hello World")
```

## Status

This software is considered ALPHA quality.  Not recommended for production
use.

## Description

This repo provides packages that let you recycle byte slices in a
persistent pool and to create a `bytes.Buffer` equivalent that dynamically
manages memory backed by a byte slice pool.

## Copyright and License

Copyright 2017 by David A. Golden. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License"). You may
obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
