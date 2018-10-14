/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	startAt int64 = 1400337722
)

var mockSnapshots = []string{
	"1400337531-1400337532-complete",
	"1400337611-1400337612-complete",
	"1400337651-1400337652-complete",
	"1400337671-1400337672-complete",
	"1400337691-1400337692-complete",
	"1400337706-1400337707-complete",
	"1400337711-1400337712-complete",
	"1400337716-1400337717-complete",
	"1400337721-1400337722-complete",
}

func mockConfig() {
	tmpRepository, err := ioutil.TempDir("", "snaprd_testing")
	if err != nil {
		panic("could not create temporary directory")
	}
	config = &Config{
		repository: tmpRepository,
		Schedule:   "testing2",
		MaxKeep:    2,
		NoPurge:    false,
		SchedFile:  "testdata/snaprd.schedules",
	}
}

func mockRepository() {
	for _, s := range mockSnapshots {
		os.MkdirAll(filepath.Join(config.repository, dataSubdir, s), 0777)
	}
}

func assertSnapshotChanLen(t *testing.T, c chan *snapshot, want int) {
	if got := len(c); got != want {
		t.Errorf("channel %v contains %v snapshots, wanted %v", c, got, want)
	}
}

func assertSnapshotChanItem(t *testing.T, c chan *snapshot, want string) {
	if got := <-c; got.String() != want {
		t.Errorf("prune() obsoleted %v, wanted %v", got.String(), want)
	}
}

type pruneTestPair struct {
	iteration time.Duration
	obsoleted []string
}

func TestPrune(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	mockConfig()
	mockRepository()
	schedules.addFromFile(config.SchedFile)
	defer os.RemoveAll(config.repository)
	cl := newSkewClock(startAt)
	c := make(chan *snapshot, 100)

	tests := []pruneTestPair{
		{0,
			[]string{},
		},
		{schedules[config.Schedule][0],
			[]string{
				"1400337706-1400337707 Obsolete",
			},
		},
		{schedules[config.Schedule][0] * 10,
			[]string{
				"1400337716-1400337717 Obsolete",
				"1400337711-1400337712 Obsolete",
				"1400337691-1400337692 Obsolete",
			},
		},
		{schedules[config.Schedule][0] * 20,
			[]string{
				"1400337531-1400337532 Obsolete",
				"1400337721-1400337722 Obsolete",
				"1400337611-1400337612 Obsolete",
				"1400337671-1400337672 Obsolete",
			},
		},
	}

	for _, pair := range tests {
		cl.forward(pair.iteration)
		prune(c, cl)
		assertSnapshotChanLen(t, c, len(pair.obsoleted))
		for _, snS := range pair.obsoleted {
			assertSnapshotChanItem(t, c, snS)
		}
	}
}
