package main

import (
    "log"
    "time"
)

var config *Config

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    var c chan string = make(chan string)
    config = LoadConfig()
    log.Println("config:", config)
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    for _, sn := range snapshots {
        log.Println(sn)
    }
    lastGood := snapshots.lastGood()
    if lastGood != nil {
        log.Println("lastgood:", lastGood)
    } else {
        log.Println("lastgood: could not find suitable base snapshot")
    }
    go CreateSnapshot(c, lastGood)
    for {
        select {
        case m := <-c:
            log.Println(m)
        case <- time.After(time.Hour*10):
            log.Println("timeout")
            return
        }
    }
}
