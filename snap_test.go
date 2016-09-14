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
	sdate    int64 = 1400268377
	edate    int64 = 1400268387
	lastGood       = "1400337721-1400337722 Complete"
)

func TestNewSnapshot(t *testing.T) {
	out := strconv.FormatInt(sdate, 10) + "-" + strconv.FormatInt(edate, 10) + " Complete"
	sn := newSnapshot(time.Unix(sdate, 0), time.Unix(edate, 0), stateComplete)
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
		os.MkdirAll(filepath.Join(config.repository, dataSubdir, s), 0777)
	}
}

type danglingTestPair struct {
	i   int
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

	sl := findDangling(cl)
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

	sl, _ := findSnapshots(cl)
	if s := sl.lastGood().String(); s != lastGood {
		t.Errorf("lastGood() found %v, should be %v", s, lastGood)
	}
	// Advance to next snapshot the is not (yet) complete, see if this is
	// omitted as it should
	os.Mkdir(filepath.Join(config.repository, dataSubdir, "1400337727-0-incomplete"), 0777)
	cl.skew -= schedules["testing2"][0]
	sl, _ = findSnapshots(cl)
	if s := sl.lastGood().String(); s != lastGood {
		t.Errorf("lastGood() found %v, should be %v", s, lastGood)
	}
}

type snStateTestPair struct {
	include snapshotState
	exclude snapshotState
	sl      *snapshotList
}

func TestSnapshotState(t *testing.T) {
	slIn := &snapshotList{
		{time.Unix(1400337531, 0), time.Unix(1400337532, 0), stateComplete},
		{time.Unix(1400337611, 0), time.Unix(1400337612, 0), stateComplete},
		{time.Unix(1400337651, 0), time.Unix(1400337652, 0), statePurging},
		{time.Unix(1400337671, 0), time.Unix(1400337672, 0), stateComplete},
		{time.Unix(1400337691, 0), time.Unix(1400337692, 0), stateComplete},
		{time.Unix(1400337706, 0), time.Unix(1400337707, 0), stateComplete},
		{time.Unix(1400337711, 0), time.Unix(1400337712, 0), stateObsolete},
		{time.Unix(1400337716, 0), time.Unix(1400337717, 0), stateComplete},
		{time.Unix(1400337721, 0), time.Unix(1400337722, 0), stateIncomplete},
	}
	tests := []snStateTestPair{
		{
			statePurging, 0, &snapshotList{
				&snapshot{time.Unix(1400337651, 0), time.Unix(1400337652, 0), statePurging},
			},
		},
		{
			statePurging + stateObsolete, 0, &snapshotList{
				&snapshot{time.Unix(1400337651, 0), time.Unix(1400337652, 0), statePurging},
				&snapshot{time.Unix(1400337711, 0), time.Unix(1400337712, 0), stateObsolete},
			},
		},
		{
			any, stateComplete, &snapshotList{
				&snapshot{time.Unix(1400337651, 0), time.Unix(1400337652, 0), statePurging},
				&snapshot{time.Unix(1400337711, 0), time.Unix(1400337712, 0), stateObsolete},
				&snapshot{time.Unix(1400337721, 0), time.Unix(1400337722, 0), stateIncomplete},
			},
		},
	}
	for _, pair := range tests {
		slOut := slIn.state(pair.include, pair.exclude)
		lslOut, lslWant := len(slOut), len(*pair.sl)
		if lslOut != lslWant {
			t.Errorf("state() delivered %v items, should be %v", lslOut, lslWant)
			// fail whole test to avoid out of range errors later
			t.FailNow()
		}
		for i := range slOut {
			sOut, sWant := slOut[i].String(), (*pair.sl)[i].String()
			if sOut != sWant {
				t.Errorf("state() found %v, should be %v", sOut, sWant)
			}
		}
	}
}

//func parseSnapshotName(s string) (time.Time, time.Time, SnapshotState, error) {

type snParseTestPair struct {
	in  string
	out *snapshot
}

func TestParseSnapshotName(t *testing.T) {
	testsGood := []snParseTestPair{
		{
			"1400337531-1400337532-complete",
			&snapshot{time.Unix(1400337531, 0), time.Unix(1400337532, 0), stateComplete},
		},
		{
			"1400337651-1400337652-purging",
			&snapshot{time.Unix(1400337651, 0), time.Unix(1400337652, 0), statePurging},
		},
		{
			"1400337721-1400337722-obsolete",
			&snapshot{time.Unix(1400337721, 0), time.Unix(1400337722, 0), stateObsolete},
		},
	}
	for _, pair := range testsGood {
		stime, etime, state, err := parseSnapshotName(pair.in)
		sOut := &snapshot{stime, etime, state}
		if err != nil {
			t.Errorf("parseSnapshotName(%v) gave error %v", pair.in, err)
		}
		if sOut.String() != pair.out.String() {
			t.Errorf("parseSnapshotName(%v) gave %v, should be %v", pair.in, sOut, pair.out)
		}
	}
	testsBad := []string{
		"1400337531-1400337532-completeXXX",
		"-1400337652-purging",
		"1400337721-1400337722-incomplete",
		"1400337721.0-1400337722-incomplete",
	}
	for _, s := range testsBad {
		_, _, _, err := parseSnapshotName(s)
		if err == nil {
			t.Errorf("parseSnapshotName(%v) did not fail, but it should", s)
		}
	}
}
