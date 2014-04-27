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

func periodic(f func(), d time.Duration) {
    ticker := time.NewTicker(d)
    for {
        <-ticker.C
        f()
    }
}

func subcmdRun() {
    // run snapshot scheduler at the lowest interval rate
    go periodic(func() {
        snapshots, err := FindSnapshots()
        if err != nil {
            log.Println(err)
        }
        //log.Println("found", len(snapshots), "snapshots in repository", config.repository)
        lastGood := snapshots.state(STATE_COMPLETE, NONE).lastGood()
        if lastGood != nil {
            //log.Println("lastgood:", lastGood)
        } else {
            log.Println("lastgood: could not find suitable base snapshot")
        }
        CreateSnapshot(lastGood)
        prune()
    }, schedules[config.schedule][0])

    go periodic(func() {
        snapshots, err := FindSnapshots()
        if err != nil {
            log.Println(err)
        }
        for _, s := range snapshots.state(STATE_OBSOLETE + STATE_PURGING, STATE_COMPLETE) {
            s.purge()
        }
    }, time.Second*3)

    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
    fmt.Println("Got signal:", <-c)
}

func subcmdList() {
    intervals := schedules[config.schedule]
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    for n := len(intervals)-2; n >= 0; n-- {
        if config.showAll {
            snapshots = snapshots.state(ANY, NONE)
        } else {
            snapshots = snapshots.state(STATE_COMPLETE, NONE)
        }
        for _, sn := range snapshots.interval(intervals, n) {
            stime := sn.startTime.Format("2006-01-02 Monday 15:04:05")
            var dur time.Duration = 0
            if sn.endTime.After(sn.startTime) {
                dur = sn.endTime.Sub(sn.startTime)
            }
            if config.verbose {
                fmt.Printf("%d %s (%s, %s, %s) \"%s\"\n", n, stime, dur, intervals[n], sn.state, sn.Name())
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
        subcmdRun()
    case "list":
        subcmdList()
    }
}
