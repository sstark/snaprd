/* See the file "LICENSE.txt" for the full license governing this code. */
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"
)

var config *Config

func Debugf(format string, args ...interface{}) {
    if os.Getenv("SNAPRD_DEBUG") == "1" {
        log.Printf("<DEBUG> "+format, args...)
    }
}

func FindDangling(q chan *Snapshot) {
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    for _, sn := range snapshots.state(STATE_OBSOLETE+STATE_PURGING, STATE_COMPLETE) {
        Debugf("found dangling snapshot: %s", sn)
        q <- sn
    }
}

func LastGoodFromDisk() *Snapshot {
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    sn := snapshots.state(STATE_COMPLETE, NONE).lastGood()
    if sn == nil {
        log.Println("lastgood: could not find suitable base snapshot")
    }
    return sn
}

// The LastGoodTicker is the clock for the create loop.
// It takes the last created snapshot on its input channel
// and outputs it on the output channel, but only after an
// appropriate waiting time.
// To start things off, the first lastGood snapshot has to
// be read from disk.
func LastGoodTicker(in, out chan *Snapshot) {
    var gap, wait time.Duration
    var sn *Snapshot
    sn = LastGoodFromDisk()
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
            gap = time.Now().Sub(sn.startTime)
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

func subcmdRun() (ferr error) {
    pl := NewPidLocker(filepath.Join(config.repository, ".pid"))
    pl.Lock()
    defer pl.Unlock()
    if !config.NoWait {
        time.Sleep(time.Second * 30)
    }
    createExit := make(chan bool)
    createExitDone := make(chan bool)
    purgeExit := make(chan bool)
    purgeExitDone := make(chan bool)
    killRsync := make(chan bool, 1)
    // The obsoleteQueue should not be larger than the absolute
    // number of expected snapshots. However, there is no way
    // (yet) to calculate that number.
    obsoleteQueue := make(chan *Snapshot, 100)
    lastGoodIn := make(chan *Snapshot)
    lastGoodOut := make(chan *Snapshot)

    go LastGoodTicker(lastGoodIn, lastGoodOut)

    // run snapshot scheduler at the lowest interval rate
    go func() {
        var lastGood *Snapshot
        var breakLoop bool
        for {
            select {
            case <-createExit:
                Debugf("gracefully exiting snapshot creation goroutine")
                breakLoop = true
            case lastGood = <-lastGoodOut:
            }
            if breakLoop {
                Debugf("breaking loop")
                createExitDone <- true
                return
            }
            sn, err := CreateSnapshot(lastGood, killRsync)
            if err != nil || sn == nil {
                Debugf("snapshot creation finally failed (%s), exit loop", err)
                breakLoop = true
                Debugf("killing purger")
                purgeExit <- true
                purgeExit = nil
                <-purgeExitDone
                purgeExitDone = nil
                go func() { createExit <- true }()
                ferr = err
            }
            lastGoodIn <- sn
            Debugf("pruning")
            prune(obsoleteQueue)
        }
    }()
    Debugf("started snapshot creation goroutine")

    // Usually the purger gets its input from the obsoleteQueue.
    // But there could be snapshots left behind from a previously
    // failed snaprd run, so we fill the obsoleteQueue once at the
    // beginning
    FindDangling(obsoleteQueue)
    go func() {
        for {
            select {
            case <-purgeExit:
                Debugf("gracefully exiting purge goroutine")
                purgeExitDone <- true
                return
            case sn := <-obsoleteQueue:
                if !config.NoPurge {
                    sn.purge()
                }
            }
        }
    }()
    Debugf("started purge goroutine")

    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
    select {
    case sig := <-sigc:
        Debugf("Got signal", sig)
        switch sig {
        case syscall.SIGINT, syscall.SIGTERM:
            log.Println("-> Immediate exit")
            killRsync <- true
            ferr = nil
        case syscall.SIGUSR1:
            log.Println("-> Graceful exit")
            if createExit != nil {
                createExit <- true
            }
            if createExitDone != nil {
                <-createExitDone
            }
            if purgeExit != nil {
                purgeExit <- true
            }
            if purgeExitDone != nil {
                <-purgeExitDone
            }
            ferr = nil
        }
    case <-createExitDone:
    }
    return
}

func subcmdList() {
    intervals := schedules[config.Schedule]
    snapshots, err := FindSnapshots()
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
        snapshots := snapshots.interval(intervals, n)
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
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    config = LoadConfig()
    if config == nil {
        log.Fatal("no config, don't know what to do!")
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
