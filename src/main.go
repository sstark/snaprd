package main

import (
    "log"
    "time"
    "fmt"
    "os"
)

var config *Config

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    var c chan error = make(chan error)
    config = LoadConfig()
    if config == nil {
        log.Fatal("no config, don't know what to do!")
    }
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    log.Println("found", len(snapshots), "snapshots")

    switch cmd {
    case "run":
        {
            lastGood := snapshots.lastGood()
            if lastGood != nil {
                log.Println("lastgood:", lastGood)
            } else {
                log.Println("lastgood: could not find suitable base snapshot")
            }
            go CreateSnapshot(c, lastGood)
            for {
                select {
                case e := <-c:
                    if e == nil {
                        log.Println("Snapshot created")
                    } else {
                        log.Println("rsync error:", e)
                    }
                case <-time.After(time.Hour * 10):
                    log.Println("timeout")
                    return
                }
            }
        }
    case "list":
        {
            for _, sn := range snapshots {
                stime := time.Unix(sn.startTime, 0).Format("2006-01-02 Monday 15:04:05")
                var dur time.Duration = 0
                if sn.endTime > sn.startTime {
                    dur = time.Duration(sn.endTime-sn.startTime) * time.Second
                }
                fmt.Printf("* %s (%s) \"%s\" S%s\n", stime, dur, sn.Name(), sn.state)
            }
            os.Exit(0)
        }
    }
}
