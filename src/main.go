package main

import (
    "log"
    "time"
    "fmt"
    "os"
    "sync"
)

var config *Config

func runLoop() {
        for {
            snapshots, err := FindSnapshots(ALL)
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
            waitTime := time.Second * time.Duration(schedules[config.schedule][0])
            log.Println("waiting for", time.Duration(waitTime))
            time.Sleep(waitTime)
        }
}

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    config = LoadConfig()
    if config == nil {
        log.Fatal("no config, don't know what to do!")
    }

    switch cmd {
    case "run": runLoop()
    case "list":
        {
            snapshots, err := FindSnapshots(ALL)
            if err != nil {
                log.Println(err)
            }
            for i, sn := range snapshots {
                stime := time.Unix(sn.startTime, 0).Format("2006-01-02 Monday 15:04:05")
                var dur time.Duration = 0
                var dist time.Duration = 0
                if sn.endTime > sn.startTime {
                    dur = time.Duration(sn.endTime-sn.startTime) * time.Second
                    if i < len(snapshots)-1 {
                        dist = time.Duration(snapshots[i+1].startTime-sn.startTime) * time.Second
                    }
                }
                fmt.Printf("* %s (%s, %s) S%s\n", stime, dur, dist, sn.state)
            }
            os.Exit(0)
        }
    case "prune": prune()
    }
}
