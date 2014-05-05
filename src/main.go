package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

var config *Config

func Debugf(format string, args ...interface{}) {
    if os.Getenv("SNAPRD_DEBUG") == "1" {
        log.Printf("<DEBUG> "+format, args...)
    }
}

func periodic(f func(), d time.Duration) {
    ticker := time.NewTicker(d)
    for {
        f()
        <-ticker.C
    }
}

func subcmdRun() {
    // run snapshot scheduler at the lowest interval rate
    go periodic(func() {
        snapshots, err := FindSnapshots()
        if err != nil {
            log.Println(err)
        }
        lastGood := snapshots.state(STATE_COMPLETE, NONE).lastGood()
        if lastGood != nil {
            Debugf("lastgood: %s\n", lastGood.String())
        } else {
            log.Println("lastgood: could not find suitable base snapshot")
        }
        CreateSnapshot(lastGood)
        prune()
    }, schedules[config.Schedule][0])

    if !config.NoPurge {
        go periodic(func() {
            snapshots, err := FindSnapshots()
            if err != nil {
                log.Println(err)
            }
            for _, s := range snapshots.state(STATE_OBSOLETE + STATE_PURGING, STATE_COMPLETE) {
                s.purge()
            }
        }, time.Second*3)
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
    fmt.Println("Got signal:", <-c)
}

func subcmdList() {
    intervals := schedules[config.Schedule]
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    for n := len(intervals)-2; n >= 0; n-- {
        Debugf("listing interval %d", n)
        if config.ShowAll {
            snapshots = snapshots.state(ANY, NONE)
        } else {
            snapshots = snapshots.state(STATE_COMPLETE, NONE)
        }
        snapshots := snapshots.interval(intervals, n)
        Debugf("snapshots in interval %d: %s", n, snapshots)
        if n < len(intervals)-2 {
            fmt.Printf("### from %s ago, %d/%d\n", intervals.offset(n+1), len(snapshots), intervals.goal(n))
        } else {
            fmt.Printf("### from past, %d\n", len(snapshots))
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
            if config.Verbose {
                fmt.Printf("%d %s (%s, %s/%s, %s) \"%s\"\n", n, stime, dur, intervals[n], dist, sn.state, sn.Name())
            } else {
                fmt.Printf("%s (%s, %s)\n", stime, dur, intervals[n])
            }
        }
    }
    os.Exit(0)
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
        log.Printf("Repository: %s, Origin: %s, Schedule: %s\n", config.Repository, config.Origin, config.Schedule)
        subcmdRun()
    case "list":
        fmt.Printf("Repository: %s, Origin: %s, Schedule: %s\n", config.Repository, config.Origin, config.Schedule)
        subcmdList()
    }
}
