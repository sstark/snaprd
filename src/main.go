package main

import (
    "log"
    "time"
    "fmt"
    "os"
    "sync"
)

var config *Config

func subcmdRun() {
        for {
            snapshots, err := FindSnapshots(ANY)
            if err != nil {
                log.Println(err)
            }
            log.Println("found", len(snapshots), "snapshots in repository", config.repository)
            lastGood := snapshots.lastGood()
            if lastGood != nil {
                log.Println("lastgood:", lastGood)
            } else {
                log.Println("lastgood: could not find suitable base snapshot")
            }
            var wg sync.WaitGroup
            wg.Add(1)
            go CreateSnapshot(&wg, lastGood)
            wg.Wait()
            for i:=0; i<len(schedules[config.schedule]); i++ {
                prune()
            }
            waitTime := schedules[config.schedule][0]
            log.Println("waiting for", waitTime)
            time.Sleep(waitTime)
        }
}

func subcmdList() {
    snapshots, err := FindSnapshots(ANY)
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
            fmt.Printf("* %s (%s, %s) S%s \"%s\"\n", stime, dur, dist, sn.state, sn.Name())
        } else {
            fmt.Printf("* %s (%s, %s) S%s\n", stime, dur, dist, sn.state)
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
    case "run": subcmdRun()
    case "list": subcmdList()
    }
}
