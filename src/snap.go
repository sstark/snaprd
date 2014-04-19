package main

import (
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "time"
    "strconv"
    "strings"
    "errors"
    "fmt"
)

type SnapshotState int

const (
    STATE_INCOMPLETE SnapshotState = 1 << iota
    STATE_COMPLETE
    STATE_OBSOLETE
    STATE_INDELETION
)

const ALL SnapshotState = STATE_INCOMPLETE + STATE_COMPLETE + STATE_OBSOLETE + STATE_INDELETION

func (st SnapshotState) String() string {
    s := ""
    if st&STATE_INCOMPLETE == STATE_INCOMPLETE {
        s += ":Incomplete"
    }
    if st&STATE_COMPLETE == STATE_COMPLETE {
        s += ":Complete"
    }
    if st&STATE_OBSOLETE == STATE_OBSOLETE {
        s += ":Obsolete"
    }
    if st&STATE_INDELETION == STATE_INDELETION {
        s += ":Indeletion"
    }
    return s
}

type Snapshot struct {
    startTime   time.Time
    endTime     time.Time
    state       SnapshotState
}

func newSnapshot(startTime, endTime time.Time, state SnapshotState) *Snapshot {
    return &Snapshot{startTime, endTime, state}
}

func newIncompleteSnapshot() *Snapshot {
    return &Snapshot{time.Now(), time.Time{}, STATE_INCOMPLETE}
}

func (s *Snapshot) String() string {
    stime := s.startTime.Unix()
    etime := s.endTime.Unix()
    return fmt.Sprintf("%d-%d S%s", stime, etime, s.state.String())
}

func (s *Snapshot) Name() (n string) {
    stime := s.startTime.Unix()
    etime := s.endTime.Unix()
    switch s.state {
    case STATE_INCOMPLETE:
        return fmt.Sprintf("%d-0-incomplete", stime)
    case STATE_COMPLETE:
        return fmt.Sprintf("%d-%d-complete", stime, etime)
    case STATE_COMPLETE | STATE_OBSOLETE:
        return fmt.Sprintf("%d-%d-complete,obsolete", stime, etime)
    }
    return ""
}

// Mark the latest snapshot for easy access.
// Do not fail if not possible since it is more important
// to continue creating new snapshots.
func tryLink(target string) {
    linkName := filepath.Join(config.repository, "latest")
    fi, err := os.Lstat(linkName)
    if err != nil {
        // link does not exist or can not be read
        log.Println(err)
    }
    if fi != nil {
        // link exists
        if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
            // link is indeed a symlink
            err = os.Remove(linkName)
            if err != nil {
                // link can not be removed
                log.Println(err)
            }
        }
    }
    err = os.Symlink(target, linkName)
    if err != nil {
        log.Println(err)
    } else {
        log.Println("symlink latest snapshot")
    }
}

func (s *Snapshot) transComplete() {
    oldName := filepath.Join(config.repository, s.Name())
    etime := time.Now()
    if etime.Before(s.startTime) {
        log.Fatal("endTime before startTime!")
    }
    // make all snapshots at least 1 second long
    if etime.Sub(s.startTime).Seconds() < 1 {
        etime = etime.Add(time.Second)
    }
    s.endTime = etime
    s.state = STATE_COMPLETE
    newName := filepath.Join(config.repository, s.Name())
    err := os.Rename(oldName, newName)
    if err != nil {
        log.Fatal(err)
    }
    tryLink(newName)
}

func (s *Snapshot) transObsolete() {
    oldName := filepath.Join(config.repository, s.Name())
    s.state = s.state | STATE_OBSOLETE
    newName := filepath.Join(config.repository, s.Name())
    err := os.Rename(oldName, newName)
    if err != nil {
        log.Fatal(err)
    }
}

type SnapshotList []*Snapshot

// find the last snapshot to use as a basis for the next one
func (sl SnapshotList) lastGood() *Snapshot {
    var t time.Time
    var ix int = -1
    for i, sn := range sl {
        if (sn.startTime.After(t)) && (sn.state == STATE_COMPLETE) {
            t = sn.startTime
            ix = i
        }
    }
    if ix == -1 {
        return nil
    }
    return sl[ix]
}

func parseSnapshotName(s string) (time.Time, time.Time, SnapshotState, error) {
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
    var state SnapshotState = 0
    stateInfo := strings.Split(sa[2], ",")
    for _, si := range stateInfo {
        switch si {
        case "complete":
            state += STATE_COMPLETE
        case "incomplete":
            state += STATE_INCOMPLETE
        case "obsolete":
            state += STATE_OBSOLETE
        case "indeletion":
            state += STATE_INDELETION
        }
    }
    // no state tags found
    if state == 0 {
        return time.Unix(stime, 0), time.Unix(etime, 0), state, errors.New("could not parse state: " + s)
    }
    return time.Unix(stime, 0), time.Unix(etime, 0), state, nil
}

func FindSnapshots(filterState SnapshotState) (SnapshotList, error) {
    snapshots := make(SnapshotList, 0, 256)
    files, err := ioutil.ReadDir(filepath.Join(config.repository, ""))
    if err != nil {
        return nil, errors.New("repository " + config.repository + " does not exist")
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
        if stime.After(time.Now()) {
            log.Println("ignoring snapshot with startTime in future:", f.Name())
            continue
        }
        sn := newSnapshot(stime, etime, state)
        if sn.state | filterState == filterState {
            snapshots = append(snapshots, sn)
        }
    }
    return snapshots, nil
}
