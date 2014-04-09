package main

import (
    "log"
    "time"
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
        case <- time.After(time.Hour*10):
            log.Println("timeout")
            return
        }
    }
}
