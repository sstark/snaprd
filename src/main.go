/* See the file "LICENSE.txt" for the full license governing this code. */

// Functions according to sub commands given on command line
// Main snapshot creation and purging loops

package main

import (
    "errors"
    "fmt"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"
)

var config *Config
var logger *log.Logger

func Debugf(format string, args ...interface{}) {
    if os.Getenv("SNAPRD_DEBUG") == "1" {
        logger.Output(2, "<DEBUG> "+fmt.Sprintf(format, args...))
    }
}

// The LastGoodTicker is the clock for the create loop. It takes the last
// created snapshot on its input channel and outputs it on the output channel,
// but only after an appropriate waiting time. To start things off, the first
// lastGood snapshot has to be read from disk.
func LastGoodTicker(in, out chan *Snapshot, cl Clock) {
    var gap, wait time.Duration
    var sn *Snapshot
    sn = LastGoodFromDisk(cl)
    if sn != nil {
        Debugf("lastgood from disk: %s\n", sn.String())
    }
    // kick off the loop
    go func() {
        in <- sn
        return
    }()
    for {
        sn := <-in
        if sn != nil {
            gap = cl.Now().Sub(sn.startTime)
            Debugf("gap: %s", gap)
            wait = schedules[config.Schedule][0] - gap
            if wait > 0 {
                log.Println("wait", wait, "before next snapshot")
                time.Sleep(wait)
            }
        }
        out <- sn
    }
}

// subcmdRun is the main, long-running routine and starts off a couple of
// helper goroutines.
func subcmdRun() (ferr error) {
    pl := NewPidLocker(filepath.Join(config.repository, ".pid"))
    pl.Lock()
    defer pl.Unlock()
    if !config.NoWait {
        sigc := make(chan os.Signal, 1)
        signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
        select {
        case <-sigc:
            return errors.New("-> Early exit")
        case <-time.After(time.Second * 30):
        }
    }
    createExit := make(chan bool)
    createExitDone := make(chan error)
    // The obsoleteQueue should not be larger than the absolute number of
    // expected snapshots. However, there is no way (yet) to calculate that
    // number.
    obsoleteQueue := make(chan *Snapshot, 100)
    lastGoodIn := make(chan *Snapshot)
    lastGoodOut := make(chan *Snapshot)

    cl := new(realClock)
    go LastGoodTicker(lastGoodIn, lastGoodOut, cl)

    // Snapshot creation loop
    go func() {
        var lastGood *Snapshot
        var createError error
    CREATE_LOOP:
        for {
            select {
            case <-createExit:
                Debugf("gracefully exiting snapshot creation goroutine")
                lastGoodOut = nil
                break CREATE_LOOP
            case lastGood = <-lastGoodOut:
                sn, err := CreateSnapshot(lastGood)
                if err != nil || sn == nil {
                    Debugf("snapshot creation finally failed (%s), exit loop", err)
                    createError = err
                    go func() { createExit <- true; return }()
                } else {
                    lastGoodIn <- sn
                    Debugf("pruning")
                    prune(obsoleteQueue, cl)
                }
            }
        }
        createExitDone <- createError
    }()
    Debugf("started snapshot creation goroutine")

    // Usually the purger gets its input only from prune(). But there
    // could be snapshots left behind from a previously failed snaprd run, so
    // we fill the obsoleteQueue once at the beginning.
    for _, sn := range FindDangling(cl) {
        obsoleteQueue <- sn
    }

    // Purger loop
    go func() {
        for {
            if sn := <-obsoleteQueue; !config.NoPurge {
                sn.purge()
            }
        }
    }()
    Debugf("started purge goroutine")

    // Global signal handling
    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
    select {
    case sig := <-sigc:
        Debugf("Got signal", sig)
        switch sig {
        case syscall.SIGINT, syscall.SIGTERM:
            log.Println("-> Immediate exit")
        case syscall.SIGUSR1:
            log.Println("-> Graceful exit")
            createExit <- true
            ferr = <-createExitDone
        }
    case ferr = <-createExitDone:
    }
    return
}

// subcmdList give the user an overview of what's in the repository.
func subcmdList() {
    intervals := schedules[config.Schedule]
    cl := new(realClock)
    snapshots, err := FindSnapshots(cl)
    if err != nil {
        log.Println(err)
    }
    for n := len(intervals) - 2; n >= 0; n-- {
        Debugf("listing interval %d", n)
        if config.showAll {
            snapshots = snapshots.state(ANY, NONE)
        } else {
            snapshots = snapshots.state(STATE_COMPLETE, NONE)
        }
        snapshots := snapshots.interval(intervals, n, cl)
        Debugf("snapshots in interval %d: %s", n, snapshots)
        if n < len(intervals)-2 {
            fmt.Printf("### From %s ago, %d/%d\n", intervals.offset(n+1), len(snapshots), intervals.goal(n))
        } else {
            if config.MaxKeep == 0 {
                fmt.Printf("### From past, %d/∞\n", len(snapshots))
            } else {
                fmt.Printf("### From past, %d/%d\n", len(snapshots), config.MaxKeep)
            }
        }
        for i, sn := range snapshots {
            stime := sn.startTime.Format("2006-01-02 Monday 15:04:05")
            var dur, dist time.Duration
            if i < len(snapshots)-1 {
                dist = snapshots[i+1].startTime.Sub(sn.startTime)
            }
            if sn.endTime.After(sn.startTime) {
                dur = sn.endTime.Sub(sn.startTime)
            }
            if config.verbose {
                fmt.Printf("%d %s (%s, %s/%s, %s) \"%s\"\n", n, stime, dur, intervals[n], dist, sn.state, sn.Name())
            } else {
                fmt.Printf("%s (%s, %s)\n", stime, dur, intervals[n])
            }
        }
    }
}

func main() {
    logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
    config = LoadConfig()
    if config == nil {
        log.Fatal("no config, don't know what to do!")
    }
    if config.NoLogDate {
        log.SetFlags(logger.Flags() - log.Ldate - log.Ltime)
        logger.SetFlags(logger.Flags() - log.Ldate - log.Ltime)
    }
    switch subcmd {
    case "run":
        log.Printf("%s started with pid %d\n", myName, os.Getpid())
        log.Printf("### Repository: %s, Origin: %s, Schedule: %s\n", config.repository, config.Origin, config.Schedule)
        err := subcmdRun()
        if err != nil {
            log.Println(err)
            os.Exit(1)
        }
    case "list":
        fmt.Printf("### Repository: %s, Origin: %s, Schedule: %s\n", config.repository, config.Origin, config.Schedule)
        subcmdList()
    }
    os.Exit(0)
}
