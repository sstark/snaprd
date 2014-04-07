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
    snapshots := FindSnapshots()
    for _, sn := range snapshots {
        log.Println(sn)
    }
    go CreateSnapshot(c)
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
