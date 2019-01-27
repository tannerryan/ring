# ring - high performance bloom filter
[![Build
Status](https://travis-ci.org/TheTannerRyan/ring.svg?branch=master)](https://travis-ci.org/TheTannerRyan/ring)
[![Go Report
Card](https://goreportcard.com/badge/github.com/thetannerryan/ring)](https://goreportcard.com/report/github.com/thetannerryan/ring)
[![GoDoc](https://godoc.org/github.com/TheTannerRyan/ring?status.svg)](https://godoc.org/github.com/TheTannerRyan/ring)
[![GitHub
license](https://img.shields.io/github/license/thetannerryan/ring.svg)](https://github.com/TheTannerRyan/ring/blob/master/LICENSE)

Package ring provides a high performance Go implementation of a bloom filter,
with an optional circular buffer, used for checking set membership.

## Usage
Please see the [godoc](https://godoc.org/github.com/TheTannerRyan/ring) for
usage. More information regaring performance and usage will be provided soon.

## Accuracy
Running `make` will perform unit tests, comparing the target false positive rate
with the actual rate. Here is a test against 1 million elements with a targeted
false positive rate of 0.1%. Tests fail if the number of false positives exceeds
the target.
```
=== RUN   TestBadParameters
--- PASS: TestBadParameters (0.00s)
=== RUN   TestReset
--- PASS: TestReset (1.09s)
=== RUN   TestData
--- PASS: TestData (8.94s)
=== RUN   TestBuffer
--- PASS: TestBuffer (9.01s)
PASS
>> Number of elements:  1000000
>> Target false positive rate:  0.001000
>> Number of false positives:  110
>> Actual false positive rate:  0.000110
>> Benchmark Add() (no buffer):   5000000              295 ns/op
>> Benchmark Test() (no buffer):   3000000             451 ns/op
>> Benchmark Add() (buffered):   5000000               296 ns/op
>> Benchmark Test() (buffered):      2000           831634 ns/op
ok      command-line-arguments  26.249s
```

## License
Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code is
governed by a BSD-style license that can be found in the LICENSE file.
