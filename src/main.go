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
        snapshots, err := FindSnapshots(STATE_COMPLETE, NONE)
        if err != nil {
            log.Println(err)
        }
        //log.Println("found", len(snapshots), "snapshots in repository", config.repository)
        lastGood := snapshots.lastGood()
        if lastGood != nil {
            //log.Println("lastgood:", lastGood)
        } else {
            log.Println("lastgood: could not find suitable base snapshot")
        }
        CreateSnapshot(lastGood)
        prune()
    }, schedules[config.schedule][0])

    go periodic(func() {
        snapshots, err := FindSnapshots(STATE_OBSOLETE + STATE_PURGING, STATE_COMPLETE)
        if err != nil {
            log.Println(err)
        }
        for _, s := range snapshots {
            s.purge()
        }
    }, time.Second*3)

    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
    fmt.Println("Got signal:", <-c)
}

func subcmdList() {
    var snapshots SnapshotList
    var err error
    if config.showAll {
        snapshots, err = FindSnapshots(ANY, NONE)
    } else {
        snapshots, err = FindSnapshots(STATE_COMPLETE, NONE)
    }
    if err != nil {
        log.Println(err)
    }
    for i, sn := range snapshots {
        stime := sn.startTime.Format("2006-01-02 Monday 15:04:05")
        var dur time.Duration = 0
        var dist time.Duration = 0
        if sn.endTime.After(sn.startTime) {
            dur = sn.endTime.Sub(sn.startTime)
            if i < len(snapshots)-1 {
                dist = snapshots[i+1].startTime.Sub(sn.startTime)
            }
        }
        if config.verbose {
            fmt.Printf("* %s (%s, %s) %s \"%s\"\n", stime, dur, dist, sn.state, sn.Name())
        } else {
            fmt.Printf("* %s (%s, %s) %s\n", stime, dur, dist, sn.state)
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
