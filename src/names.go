package main

import (
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "time"
    "strconv"
    "strings"
)

type SnapshotState int

const (
    STATE_INCOMPLETE SnapshotState = 1 << iota
    STATE_COMPLETE
    STATE_OBSOLETE
    STATE_INDELETION
)

func (st SnapshotState) String() string {
    s := ""
    if st & STATE_INCOMPLETE == STATE_INCOMPLETE {
        s += ":Incomplete"
    }
    if st & STATE_COMPLETE == STATE_COMPLETE {
        s += ":Complete"
    }
    if st & STATE_OBSOLETE == STATE_OBSOLETE {
        s += ":Obsolete"
    }
    if st & STATE_INDELETION == STATE_INDELETION {
        s += ":Indeletion"
    }
    return s
}

type Snapshot struct {
    startTime int64
    endTime int64
    state SnapshotState
}

func unixTimestamp() int64 {
    return time.Now().Unix()
}

func newSnapshot(startTime int64, endTime int64, state SnapshotState) *Snapshot {
    return &Snapshot{startTime, endTime, state}
}

func newIncompleteSnapshot() *Snapshot {
    return &Snapshot{unixTimestamp(), 0, STATE_INCOMPLETE}
}

func (s *Snapshot) String() string {
    stime := strconv.FormatInt(s.startTime, 10)
    etime := strconv.FormatInt(s.endTime, 10)
    return stime + "-" + etime + " S" + s.state.String()
}

func (s *Snapshot) Name() (n string) {
    stime := strconv.FormatInt(s.startTime, 10)
    etime := strconv.FormatInt(s.endTime, 10)
    switch s.state {
    case STATE_INCOMPLETE:
        return stime + "-" + "0" + "-incomplete"
    case STATE_COMPLETE:
        return stime + "-" + etime + "-complete"
    case STATE_COMPLETE | STATE_OBSOLETE:
        return stime + "-" + etime + "-complete,obsolete"
    }
    return ""
}

func (s *Snapshot) transComplete() {
    oldName := filepath.Join(config.dstPath, s.Name())
    // the +10 is only for testing!
    s.endTime = unixTimestamp()+10
    s.state = STATE_COMPLETE
    os.Rename(oldName, filepath.Join(config.dstPath, s.Name()))
}

type SnapshotList []*Snapshot

func isSnapshot(f os.FileInfo) bool {
    if !f.IsDir() {
        return false
    }
    // number-number OR
    // number-
    return true
}

func parseSnapshotName(s string) (int64, int64, SnapshotState) {
    sa := strings.Split(s, "-")
    stime, err := strconv.ParseInt(sa[0], 10, 64)
    if err != nil {
        log.Panic(err)
    }
    etime, err := strconv.ParseInt(sa[1], 10, 64)
    if err != nil {
        log.Panic(err)
    }
    var state SnapshotState = 0
    stateInfo := strings.Split(sa[2], ",")
    for _, s := range stateInfo {
        switch s {
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
    return stime, etime, state
}

func FindSnapshots() SnapshotList {
    snapshots := make(SnapshotList, 0, 256)
    files, err := ioutil.ReadDir(filepath.Join(config.dstPath, ""))
    if err != nil {
        log.Panic(err)
    }
    for _, f := range files {
        if isSnapshot(f) {
            stime, etime, state := parseSnapshotName(f.Name())
            s := newSnapshot(stime, etime, state)
            snapshots = append(snapshots, s)
        } else {
            log.Println(f.Name() + " is not a snapshot")
        }
    }
    return snapshots
}
