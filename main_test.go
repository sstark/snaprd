package main

import "os"

func ExampleSubcmdList() {
	mockConfig()
	mockRepository()
	defer os.RemoveAll(config.repository)
	cl := newSkewClock(startAt)
	subcmdList(cl)
	// Output:
	// ### From past, 1/2
	// 2014-05-17 Saturday 16:38:51 (1s, 1m20s)
	// ### From 2m20s ago, 2/2
	// 2014-05-17 Saturday 16:40:11 (1s, 40s)
	// 2014-05-17 Saturday 16:40:51 (1s, 40s)
	// ### From 1m0s ago, 2/2
	// 2014-05-17 Saturday 16:41:11 (1s, 20s)
	// 2014-05-17 Saturday 16:41:31 (1s, 20s)
	// ### From 20s ago, 4/4
	// 2014-05-17 Saturday 16:41:46 (1s, 5s)
	// 2014-05-17 Saturday 16:41:51 (1s, 5s)
	// 2014-05-17 Saturday 16:41:56 (1s, 5s)
	// 2014-05-17 Saturday 16:42:01 (1s, 5s)
}

func ExampleScheds() {
	schedules.list()
	// Output:
	// longterm: [6h0m0s 24h0m0s 168h0m0s 672h0m0s 876000h0m0s]
	// shortterm: [10m0s 2h0m0s 24h0m0s 168h0m0s 672h0m0s 876000h0m0s]
	// test1: [24h0m0s 168h0m0s 672h0m0s 876000h0m0s]
	// testing: [5s 20s 2m20s 4m40s 876000h0m0s]
	// testing2: [5s 20s 40s 1m20s 876000h0m0s]
}
