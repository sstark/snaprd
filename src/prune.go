package main

import (
    "log"
    "time"
)

func findSnapshotsInInterval(after, before time.Time) SnapshotList {
    allSnapshots, err := FindSnapshots(ANY - STATE_OBSOLETE - STATE_INCOMPLETE)
    if err != nil {
        log.Println(err)
    }
    snapshotInterval := make(SnapshotList, 0, 256)
    for _, sn := range allSnapshots {
        if sn.startTime.After(after) && sn.startTime.Before(before) {
            snapshotInterval = append(snapshotInterval, sn)
        }
    }
    return snapshotInterval
}

// TODO should add more checks:
// - don't delete hardlink base
func prune() {
    intervals := schedules[config.schedule]
    // interval 0 does not need pruning, start with 1
    for i := 1; i < len(intervals)-1; i++ {
        t := time.Now()
        for j := 0; j <= i; j++ {
            t = t.Add(-intervals[j])
        }
        iv := findSnapshotsInInterval(t.Add(-intervals[i+1]), t)
        if len(iv) > 2 {
            // last in list is youngest
            youngest := len(iv)-1
            secondYoungest := youngest-1
            dist := iv[youngest].startTime.Sub(iv[secondYoungest].startTime)
            if dist.Seconds() < intervals[i].Seconds() {
                log.Printf("mark as obsolete: %s", iv[youngest].Name())
                iv[youngest].transObsolete()
            }
        }
    }
}
