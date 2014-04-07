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
    etime := unixTimestamp()
    if etime < s.startTime {
        log.Fatal("endTime before startTime!")
    }
    // make all snapshots at least 1 second long
    if etime == s.startTime {
        etime += 1
    }
    s.endTime = etime
    s.state = STATE_COMPLETE
    os.Rename(oldName, filepath.Join(config.dstPath, s.Name()))
}

type SnapshotList []*Snapshot

// find the last snapshot to use as a basis for the next one
func (sl SnapshotList) lastGood() *Snapshot {
    var t int64 = 0
    var ix int = -1
    for i, sn := range sl {
        if (sn.startTime > t) && (sn.state == STATE_COMPLETE) {
            t = sn.startTime
            ix = i
        }
    }
    if ix == -1  {
        return nil
    }
    return sl[ix]
}

func parseSnapshotName(s string) (int64, int64, SnapshotState, error) {
    sa := strings.Split(s, "-")
    if len(sa) != 3 {
        return 0, 0, 0, errors.New("malformed snapshot name: " + s)
    }
    stime, err := strconv.ParseInt(sa[0], 10, 64)
    if err != nil {
        return 0, 0, 0, err
    }
    etime, err := strconv.ParseInt(sa[1], 10, 64)
    if err != nil {
        return 0, 0, 0, err
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
        return stime, etime, state, errors.New("could not parse state: " + s)
    }
    return stime, etime, state, nil
}

func FindSnapshots() SnapshotList {
    snapshots := make(SnapshotList, 0, 256)
    files, err := ioutil.ReadDir(filepath.Join(config.dstPath, ""))
    if err != nil {
        log.Fatal(err)
    }
    for _, f := range files {
        // normal files are allowed but ignored
        if f.IsDir() {
            stime, etime, state, err := parseSnapshotName(f.Name())
            if err != nil {
                log.Println(err)
                continue
            }
            sn := newSnapshot(stime, etime, state)
            snapshots = append(snapshots, sn)
        }
    }
    return snapshots
}
