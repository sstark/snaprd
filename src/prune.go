package main

import (
    "log"
)

// TODO should add more checks:
// - don't delete hardlink base
func prune() {
    intervals := schedules[config.schedule]
    // interval 0 does not need pruning, start with 1
    for i := 1; i < len(intervals)-1; i++ {
        snapshots, err := FindSnapshots()
        if err != nil {
            log.Println(err)
            return
        }
        if len(snapshots) < 2 {
            log.Println("less than 2 snapshots found, not pruning")
            return
        }
        iv := snapshots.interval(intervals, i).state(STATE_COMPLETE, STATE_OBSOLETE)
        pruneAgain := 0
        if len(iv) > 2 {
            if i == len(intervals)-2 {
                // highest interval pruning
                if (len(iv) > config.maxKeep) && (config.maxKeep != 0) {
                    log.Printf("%d snapshots in oldest interval", len(iv))
                    log.Printf("mark oldest as obsolete: %s", iv[0])
                    iv[0].transObsolete()
                    pruneAgain += 1
                }
            } else {
                // last in list is youngest
                youngest := len(iv) - 1
                secondYoungest := youngest - 1
                dist := iv[youngest].startTime.Sub(iv[secondYoungest].startTime)
                if (dist.Seconds() < intervals[i].Seconds()) {
                    log.Printf("mark as obsolete: %s", iv[youngest].Name())
                    iv[youngest].transObsolete()
                    pruneAgain += 1
                }
            }
            if pruneAgain > 0 { prune() }
        }
    }
}
