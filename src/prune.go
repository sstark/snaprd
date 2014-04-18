package main

import (
    "log"
    "time"
)

const hour int64 = 3600
const day int64 = hour*24
const week int64 = day*7
const month int64 = week*4 //month == 4 weeks
const future int64 = 9999999999 //the date this program will stop working

/*
  The lowest interval will always be given by how often runLoop() is
  running rsync. It should be significantly smaller than a day, everything
  between 1 and 6 hours should be fine.

  This smallest interval does not appear in the interval array since the
  snapshot in that interval won't ever be touched by prune(). Instead, those
  snapshots that overflow the smallest intervals span, (which is a day), belong
  to the next interval by definition.

  The span of an interval is always the snapshot distance of the next interval.
*/
var schedules = map[string][]int64{
    "longterm": {hour*6, day, week, month, future},
    "testing": {5, 20, 140, 560, future},
}

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
