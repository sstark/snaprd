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

// return duration long enough to stay in normal snapshot interval
func GetGroove() time.Duration {
    snapshots, err := FindSnapshots()
    if err != nil {
        return 0
    }
    lastGood := snapshots.state(STATE_COMPLETE, NONE).lastGood()
    if lastGood == nil {
        return 0
    }
    gap := time.Now().Sub(lastGood.startTime)
    Debugf("gap: %s", gap)
    wait := schedules[config.Schedule][0] - gap
    if wait > 0 {
        log.Println("wait", wait, "before next snapshot")
        return wait
    }
    return 0
}

func subcmdRun() {
    killRsync := make(chan bool, 1)
    // run snapshot scheduler at the lowest interval rate
    time.AfterFunc(GetGroove(), func() {
        ticker := time.NewTicker(schedules[config.Schedule][0])
        for {
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
            err = CreateSnapshot(lastGood, killRsync)
            if err != nil {
                Debugf("snapshot creation finally failed, exit loop")
                break
            }
            prune()
            <-ticker.C
        }
    })
    Debugf("started snapshot creation goroutine")

    if !config.NoPurge {
        ticker := time.NewTicker(schedules[config.Schedule][0] / 2)
        go func() {
            for {
                snapshots, err := FindSnapshots()
                if err != nil {
                    log.Println(err)
                }
                Debugf("purging")
                for _, s := range snapshots.state(STATE_OBSOLETE+STATE_PURGING, STATE_COMPLETE) {
                    s.purge()
                }
                <-ticker.C
            }
        }()
    }
    Debugf("started purge goroutine")

    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
    log.Println("Got signal", <-c, "-> cleaning up before exit")
    killRsync <- true
    os.Exit(0)
}

func subcmdList() {
    intervals := schedules[config.Schedule]
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
    }
    for n := len(intervals) - 2; n >= 0; n-- {
        Debugf("listing interval %d", n)
        if config.showAll {
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
            if config.verbose {
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
        log.Printf("Repository: %s, Origin: %s, Schedule: %s\n", config.repository, config.Origin, config.Schedule)
        subcmdRun()
    case "list":
        fmt.Printf("Repository: %s, Origin: %s, Schedule: %s\n", config.repository, config.Origin, config.Schedule)
        subcmdList()
    }
}
