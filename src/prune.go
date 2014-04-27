package main

import (
    "log"
)

// TODO should add more checks:
// - don't delete hardlink base
func prune() {
    intervals := schedules[config.schedule]
    snapshots, err := FindSnapshots()
    if err != nil {
        log.Println(err)
        return
    }
    // interval 0 does not need pruning, start with 1
    for i := 1; i < len(intervals)-1; i++ {
        iv := snapshots.interval(intervals, i).state(STATE_COMPLETE, STATE_OBSOLETE)
        if len(iv) > 2 {
            // last in list is youngest
            youngest := len(iv) - 1
            secondYoungest := youngest - 1
            dist := iv[youngest].startTime.Sub(iv[secondYoungest].startTime)
            if dist.Seconds() < intervals[i].Seconds() {
                log.Printf("mark as obsolete: %s", iv[youngest].Name())
                iv[youngest].transObsolete()
                // prune as often as needed
                prune()
            }
        }
    }
}
