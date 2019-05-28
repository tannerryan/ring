# ring - high performance bloom filter
[![Build
Status](https://img.shields.io/travis/TheTannerRyan/ring.svg?style=flat-square)](https://travis-ci.org/TheTannerRyan/ring)
[![codecov](https://img.shields.io/codecov/c/github/TheTannerRyan/ring.svg?style=flat-square)](https://codecov.io/gh/TheTannerRyan/ring)
[![Go Report
Card](https://goreportcard.com/badge/github.com/thetannerryan/ring?style=flat-square)](https://goreportcard.com/report/github.com/thetannerryan/ring)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5673AF.svg?style=flat-square)](https://godoc.org/github.com/TheTannerRyan/ring)
[![GitHub
license](https://img.shields.io/github/license/TheTannerRyan/ring.svg?style=flat-square)](https://github.com/TheTannerRyan/ring/blob/master/LICENSE)
[![Mentioned in Awesome
Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)

Package ring provides a high performance and thread safe Go implementation of a
bloom filter.

## Usage
Please see the [godoc](https://godoc.org/github.com/TheTannerRyan/ring) for
usage.

## Accuracy
Running `make` will perform unit tests, comparing the target false positive rate
with the actual rate. Here is a test against 1 million elements with a targeted
false positive rate of 0.1%. Tests fail if the number of false positives exceeds
the target.
```
=== RUN   TestBadParameters
--- PASS: TestBadParameters (0.00s)
=== RUN   TestReset
--- PASS: TestReset (0.26s)
=== RUN   TestData
--- PASS: TestData (14.07s)
=== RUN   TestMerge
--- PASS: TestMerge (13.78s)
=== RUN   TestMarshal
--- PASS: TestMarshal (14.48s)
PASS
>> Number of elements:  1000000
>> Target false positive rate:  0.001000
>> Number of false positives:  99
>> Actual false positive rate:  0.000099
>> Number of false negatives:  0
>> Actual false negative rate:  0.000000
>> Benchmark Add():  10000000          158 ns/op
>> Benchmark Test():  10000000         173 ns/op
ok      command-line-arguments  47.914s
```

## License
Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code is
governed by a BSD-style license that can be found in the LICENSE file.
