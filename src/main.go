package main

import (
    "log"
    "time"
)

var config *Config

func main() {
    var c chan string = make(chan string)
    config = LoadConfig()
    log.Println("config:", config)
    sl := FindSnapshots()
    for _, s := range sl {
        log.Println(s)
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
