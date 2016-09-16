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
