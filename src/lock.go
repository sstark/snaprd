package main

import (
    "log"
    "io/ioutil"
    "os"
    "strconv"
)

type PidLocker struct {
    pid int
    f string
}

func NewPidLocker(lockfile string) *PidLocker {
    return &PidLocker{
        pid: os.Getpid(),
        f: lockfile,
    }
}

func (pl *PidLocker) Lock() {
    _, err := os.Stat(pl.f)
    if err == nil {
        log.Fatalf("pid file %s already exists. Is snaprd running already?", pl.f)
    }
    Debugf("write pid %d to pidfile %s", pl.pid, pl.f)
    err = ioutil.WriteFile(pl.f, []byte(strconv.Itoa(pl.pid)), 0666)
    if err != nil {
        log.Fatalf("could not write pid file %s", pl.f)
    }
}

func (pl *PidLocker) Unlock() {
    Debugf("delete pidfile %s", pl.f)
    err := os.Remove(pl.f)
    if err != nil {
        log.Fatalf("could not remove pid file %s", pl.f)
    }
}
