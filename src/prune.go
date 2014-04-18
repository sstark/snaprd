package main

import (
    "log"
    "time"
)

// return all snapshots where stime between low and high
// TODO make sure we get a sorted list, don't rely on file system!
func findSnapshotsInInterval(low, high int64) SnapshotList {
    allSnapshots, err := FindSnapshots(ALL - STATE_OBSOLETE)
    if err != nil {
        log.Println(err)
    }
    snapshotInterval := make(SnapshotList, 0, 256)
    now := time.Now().Unix()
    for _, sn := range allSnapshots {
        if (sn.startTime < now-low) && (sn.startTime > now-high) {
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
        var low int64
        for j := 0; j <= i; j++ {
            low += intervals[j]
        }
        iv := findSnapshotsInInterval(low, low+intervals[i+1])
        if len(iv) > 2 {
            // last in list is youngest
            yi := len(iv)-1
            dist := iv[yi].startTime - iv[yi-1].startTime
            if dist < intervals[i] {
                stime := time.Unix(iv[yi].startTime, 0).Format("2006-01-02 Monday 15:04:05")
                log.Printf("Mark as obsolete: %s \"%s\"\n", stime, iv[yi].Name())
                //Mark obsolete. This also makes sure that during the next run
                //of prune() another snapshot will be seen as the youngest.
                iv[yi].transObsolete()
            }
        }
    }
}
