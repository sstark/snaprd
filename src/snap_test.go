/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
    "os"
    "path/filepath"
    "strconv"
    "testing"
    "time"
)

const (
    sdate int64 = 1400268377
    edate int64 = 1400268387
    lastGood = "1400337721-1400337722 Complete"
)

func TestNewSnapshot(t *testing.T) {
    out := strconv.FormatInt(sdate, 10) + "-" + strconv.FormatInt(edate, 10) + " Complete"
    sn := newSnapshot(time.Unix(sdate, 0), time.Unix(edate, 0), STATE_COMPLETE)
    if s := sn.String(); s != out {
        t.Errorf("sn.String() = %v, want %v", s, out)
    }
}

var mockSnapshotsDangling = []string{
    "1400337531-1400337532-complete",
    "1400337611-1400337612-complete",
    "1400337651-1400337652-purging",
    "1400337671-1400337672-complete",
    "1400337691-1400337692-complete",
    "1400337706-1400337707-complete",
    "1400337711-1400337712-obsolete",
    "1400337716-1400337717-complete",
    "1400337721-1400337722-complete",
}

func mockRepositoryDangling() {
    for _, s := range mockSnapshotsDangling {
        os.MkdirAll(filepath.Join(config.repository, DATA_SUBDIR, s), 0777)
    }
}

type danglingTestPair struct {
    i int
    snS string
}

func TestFindDangling(t *testing.T) {
    var tests = []danglingTestPair{
        {0, "1400337651-1400337652 Purging"},
        {1, "1400337711-1400337712 Obsolete"},
    }
    mockConfig()
    mockRepositoryDangling()
    defer os.RemoveAll(config.repository)
    cl := newSkewClock(startAt)

    sl := FindDangling(cl)
    lgot, lwant := len(sl), len(tests)
    if lgot != lwant {
        t.Errorf("FindDangling() found %v, should be %v", lgot, lwant)
    }
    for _, pair := range tests {
        if s := sl[pair.i].String(); s != pair.snS {
            t.Errorf("FindDangling found %v, should be %v", s, pair.snS)
        }
    }
}

func TestLastGood(t *testing.T) {
    mockConfig()
    mockRepositoryDangling()
    defer os.RemoveAll(config.repository)
    cl := newSkewClock(startAt)

    sl, _ := FindSnapshots(cl)
    if s := sl.lastGood().String(); s != lastGood {
        t.Errorf("lastGood() found %v, should be %v", s, lastGood)
    }
    // Advance to next snapshot the is not (yet) complete, see if this is
    // omitted as it should
    os.Mkdir(filepath.Join(config.repository, DATA_SUBDIR, "1400337727-0-incomplete"), 0777)
    cl.skew -= schedules["testing2"][0]
    sl, _ = FindSnapshots(cl)
    if s := sl.lastGood().String(); s != lastGood {
        t.Errorf("lastGood() found %v, should be %v", s, lastGood)
    }
}
