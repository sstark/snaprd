package main

import (
    "log"
    "time"
    "fmt"
)

const day int64 = 86400
const week int64 = 604800
const month int64 = 2419200 //month == 4 weeks
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
var intervals = [...]int64{day, week, month, future}

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
    for i := 0; i < len(intervals)-1; i++ {
        log.Println("pruning interval", i)
        interval := findSnapshotsInInterval(intervals[i], intervals[i+1])
        if len(interval) > 1 {
            // last in list is youngest
            yi := len(interval)-1
            if (interval[yi-1].startTime - interval[yi].startTime) < day {
                stime := time.Unix(interval[yi].startTime, 0).Format("2006-01-02 Monday 15:04:05")
                fmt.Printf("Mark as obsolete: %s \"%s\"\n", stime, interval[yi].Name())
                //Mark obsolete. This also makes sure that during the next run
                //of prune() another snapshot will be seen as the youngest.
                interval[yi].transObsolete()
            }
        }
    }
}
