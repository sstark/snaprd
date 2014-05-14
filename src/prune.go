/* See the file "LICENSE.txt" for the full license governing this code. */
package main

import (
    "log"
)

// Sieves snapshots according to schedule and marks
// them as obsolete. Also, enqueue them in the buffered
// channel q for later reuse or deletion
func prune(q chan *Snapshot) {
    intervals := schedules[config.Schedule]
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
        pruneAgain := false
        if len(iv) > 2 {
            // prune highest interval by maximum number
            if (i == len(intervals)-2) &&
                (len(iv) > config.MaxKeep) &&
                (config.MaxKeep != 0) {
                Debugf("%d snapshots in oldest interval", len(iv))
                log.Printf("mark oldest as obsolete: %s", iv[0])
                iv[0].transObsolete()
                pruneAgain = true
            }
            // regularly prune by sieving
            youngest := len(iv) - 1
            secondYoungest := youngest - 1
            dist := iv[youngest].startTime.Sub(iv[secondYoungest].startTime)
            if dist.Seconds() < intervals[i].Seconds() {
                log.Printf("mark as obsolete: %s", iv[youngest].Name())
                iv[youngest].transObsolete()
                q <- iv[youngest]
                pruneAgain = true
            }
            if pruneAgain {
                prune(q)
            }
        }
    }
}
