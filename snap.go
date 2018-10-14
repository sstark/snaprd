/* See the file "LICENSE.txt" for the full license governing this code. */

// Snapshot in memory and on-disk format

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type snapshotState uint

const none snapshotState = 0
const (
	stateIncomplete snapshotState = 1 << iota
	stateComplete
	stateObsolete
	statePurging
	any = (1 << iota) - 1
)

func (st snapshotState) String() string {
	switch st {
	case stateIncomplete:
		return "Incomplete"
	case stateComplete:
		return "Complete"
	case stateObsolete:
		return "Obsolete"
	case statePurging:
		return "Purging"
	}
	return "Unknown"
}

type snapshot struct {
	startTime time.Time
	endTime   time.Time
	state     snapshotState
}

func newSnapshot(startTime, endTime time.Time, state snapshotState) *snapshot {
	return &snapshot{startTime, endTime, state}
}

func newIncompleteSnapshot(cl clock) *snapshot {
	return &snapshot{cl.Now(), time.Time{}, stateIncomplete}
}

func (s *snapshot) String() string {
	stime := s.startTime.Unix()
	etime := s.endTime.Unix()
	return fmt.Sprintf("%d-%d %s", stime, etime, s.state.String())
}

// Name returns the relative pathname for the receiver snapshot.
func (s *snapshot) Name() string {
	stime := s.startTime.Unix()
	etime := s.endTime.Unix()
	switch s.state {
	case stateIncomplete:
		return fmt.Sprintf("%d-0-incomplete", stime)
	case stateComplete:
		return fmt.Sprintf("%d-%d-complete", stime, etime)
	case stateObsolete:
		return fmt.Sprintf("%d-%d-obsolete", stime, etime)
	case statePurging:
		return fmt.Sprintf("%d-%d-purging", stime, etime)
	}
	return fmt.Sprintf("%d-%d-unknown", stime, etime)
}

// FullName returns the full pathname for the receiver snapshot.
func (s *snapshot) FullName() string {
	return filepath.Join(config.repository, dataSubdir, s.Name())
}

// transComplete transitions the receiver to complete state.
func (s *snapshot) transComplete(cl clock) {
	oldName := s.FullName()
	etime := cl.Now()
	if etime.Before(s.startTime) {
		log.Fatal("endTime before startTime!")
	}
	// make all snapshots at least 1 second long
	if etime.Sub(s.startTime).Seconds() < 1 {
		etime = etime.Add(time.Second)
	}
	s.endTime = etime
	s.state = stateComplete
	newName := s.FullName()
	debugf("renaming complete snapshot %s -> %s", oldName, newName)
	if oldName != newName {
		err := os.Rename(oldName, newName)
		if err != nil {
			log.Fatal(err)
		}
	}
	updateSymlinks()
	overwriteSymlink(filepath.Join(dataSubdir, s.Name()), filepath.Join(config.repository, "latest"))
}

// transObsolete transitions the receiver to obsolete state.
func (s *snapshot) transObsolete() {
	oldName := s.FullName()
	s.state = stateObsolete
	newName := s.FullName()
	if oldName != newName {
		err := os.Rename(oldName, newName)
		if err != nil {
			log.Fatal(err)
		}
	}
	updateSymlinks()
}

// transPurging transitions the receiver to purging state.
func (s *snapshot) transPurging() {
	oldName := s.FullName()
	s.state = statePurging
	newName := s.FullName()
	if oldName != newName {
		err := os.Rename(oldName, newName)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// transIncomplete generates a new incomplete snapshot based on a previous one.
// Can be used to try to use previous incomplete snapshots, or even to reuse
// obsolete ones.
func (s *snapshot) transIncomplete(cl clock) {
	oldName := s.FullName()
	s.startTime = cl.Now()
	s.endTime = time.Time{}
	s.state = stateIncomplete
	newName := s.FullName()
	debugf("renaming incomplete snapshot %s -> %s", oldName, newName)
	if oldName != newName {
		err := os.Rename(oldName, newName)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// purge deletes the receiver snapshot from disk.
func (s *snapshot) purge() {
	s.transPurging()
	path := s.FullName()
	log.Println("purging", s.Name())
	err := os.RemoveAll(path)
	if err != nil {
		log.Printf("error when purging \"%s\" (ignored): %s", s.Name(), err)
	}
	log.Println("finished purging", s.Name())
}

func (s *snapshot) matchFilter(f snapshotState) bool {
	//log.Println("filter:", strconv.FormatInt(int64(s.state), 2), strconv.FormatInt(int64(f), 2), strconv.FormatBool(s.state & f == s.state))
	//log.Println(strconv.FormatInt(int64(any), 2))
	return (s.state & f) == s.state
}

type snapshotList []*snapshot

// Find the last snapshot to use as a basis for the next one.
func (sl snapshotList) lastGood() *snapshot {
	var t time.Time
	var ix = -1
	for i, sn := range sl {
		if (sn.startTime.After(t)) && (sn.state == stateComplete) {
			t = sn.startTime
			ix = i
		}
	}
	if ix == -1 {
		return nil
	}
	return sl[ix]
}

// Find the last snapshot in a given list.
func (sl snapshotList) last() *snapshot {
	var t time.Time
	var ix = -1
	for i, sn := range sl {
		if sn.startTime.After(t) {
			t = sn.startTime
			ix = i
		}
	}
	if ix == -1 {
		return nil
	}
	return sl[ix]
}

// parseSnapshotName split the given string up into the various values needed
// for creating a Snapshot struct.
func parseSnapshotName(s string) (time.Time, time.Time, snapshotState, error) {
	sa := strings.Split(s, "-")
	var zero time.Time
	if len(sa) != 3 {
		return zero, zero, 0, errors.New("malformed snapshot name: " + s)
	}
	stime, err := strconv.ParseInt(sa[0], 10, 64)
	if err != nil {
		return zero, zero, 0, err
	}
	etime, err := strconv.ParseInt(sa[1], 10, 64)
	if err != nil {
		return zero, zero, 0, err
	}
	var state snapshotState
	switch sa[2] {
	case "complete":
		state = stateComplete
	case "incomplete":
		state = stateIncomplete
	case "obsolete":
		state = stateObsolete
	case "purging":
		state = statePurging
	}
	if state == 0 {
		return time.Unix(stime, 0), time.Unix(etime, 0), state, errors.New("could not parse state: " + s)
	}
	if state == stateIncomplete && etime != 0 {
		return zero, zero, 0, errors.New("incomplete state but non-zero end time: " + s)
	}
	return time.Unix(stime, 0), time.Unix(etime, 0), state, nil
}

type snapshotListByStartTime snapshotList

func (sl snapshotListByStartTime) Len() int {
	return len(sl)
}
func (sl snapshotListByStartTime) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}
func (sl snapshotListByStartTime) Less(i, j int) bool {
	return sl[i].startTime.Before(sl[j].startTime)
}

// findSnapshots() reads the repository directory and returns a list of
// Snapshot pointers for all valid snapshots it could find.
func findSnapshots(cl clock) (snapshotList, error) {
	snapshots := make(snapshotList, 0, 256)
	dataPath := filepath.Join(config.repository, dataSubdir, "")
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		return nil, errors.New("Repository " + dataPath + " does not exist")
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		stime, etime, state, err := parseSnapshotName(f.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		if stime.After(cl.Now()) {
			log.Println("ignoring snapshot with startTime in future:", f.Name())
			continue
		}
		sn := newSnapshot(stime, etime, state)
		snapshots = append(snapshots, sn)
	}
	sort.Sort(snapshotListByStartTime(snapshots))
	return snapshots, nil
}

// Return a new list of snapshots that fall into the given time period.
func (sl snapshotList) period(after, before time.Time) snapshotList {
	slNew := make(snapshotList, 0, len(sl))
	for _, sn := range sl {
		if sn.startTime.After(after) && sn.startTime.Before(before) {
			slNew = append(slNew, sn)
		}
	}
	return slNew
}

// Return a list of snapshots within the given interval.
func (sl snapshotList) interval(intervals intervalList, i int, cl clock) snapshotList {
	t := cl.Now()
	from := t.Add(-intervals.offset(i + 1))
	to := t.Add(-intervals.offset(i))
	return sl.period(from, to)
}

// Return a filtered list of snapshots that match (include) or don't match
// (exclude) the given state mask.
func (sl snapshotList) state(include, exclude snapshotState) snapshotList {
	slNew := make(snapshotList, 0, len(sl))
	for _, sn := range sl {
		if sn.matchFilter(include) && sn.matchFilter(^exclude) {
			slNew = append(slNew, sn)
		}
	}
	return slNew
}

// findDangling returns a list of obsolete or purged snapshots.
func findDangling(cl clock) snapshotList {
	snapshots, err := findSnapshots(cl)
	if err != nil {
		log.Println(err)
	}
	slNew := make(snapshotList, 0, len(snapshots))
	for _, sn := range snapshots.state(stateObsolete+statePurging, stateComplete) {
		debugf("found dangling snapshot: %s", sn)
		slNew = append(slNew, sn)
	}
	return slNew
}

// lastGoodFromDisk lists the snapshots in the repository and returns a pointer
// to the youngest complete snapshot.
func lastGoodFromDisk(cl clock) *snapshot {
	snapshots, err := findSnapshots(cl)
	if err != nil {
		log.Println(err)
	}
	sn := snapshots.state(stateComplete, none).lastGood()
	if sn == nil {
		log.Println("lastgood: could not find suitable base snapshot")
	}
	return sn
}

// lastIncompleteFromDisk lists the snapshots in the repository and returns a pointer
// to the youngest incomplete snapshot, for possible reuse.
func lastReusableFromDisk(cl clock) *snapshot {
	snapshots, err := findSnapshots(cl)
	if err != nil {
		log.Println(err)
	}
	sn := snapshots.state(stateIncomplete, none).last()
	return sn
}
