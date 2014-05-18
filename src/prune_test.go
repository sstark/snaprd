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
    startAt  int64 = 1400337722
    prunedSn       = "1400337706-1400337707 Obsolete"
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

type skewClock struct {
    skew time.Duration
}

func (cl *skewClock) Now() time.Time {
    return time.Now().Add(-cl.skew)
}

func newSkewClock(i int64) *skewClock {
    d := time.Now().Sub(time.Unix(i, 0))
    return &skewClock{skew: d}
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
    }
}

func mockRepository() {
    for _, s := range mockSnapshots {
        os.MkdirAll(filepath.Join(config.repository, DATA_SUBDIR, s), 0777)
    }
}

func TestPrune(t *testing.T) {
    log.SetOutput(ioutil.Discard)
    mockConfig()
    mockRepository()
    defer os.RemoveAll(config.repository)
    // A clock from the time the last snapshot in the
    // mock repository was created
    cl := newSkewClock(startAt)
    c := make(chan *Snapshot, 100)
    prune(c, cl)
    if len(c) > 0 {
        t.Errorf("prune() obsoleted %v, channel should be empty", <-c)
    }
    // fast forward the time by one snapshot
    cl.skew -= schedules[config.Schedule][0]
    prune(c, cl)
    if s := <-c; s.String() != prunedSn {
        t.Errorf("prune() obsoleted %v, wanted %v", s, prunedSn)
    }
}
