package main

import (
    "strings"
    "time"
)

const (
    second = time.Second
    minute = time.Minute
    hour   = time.Hour
    day    = hour * 24
    week   = day * 7
    month  = week * 4
    year   = day * 365
    long   = year * 100
)

type intervalList []time.Duration

// returns how long ago the given interval started
func (il intervalList) offset(i int) time.Duration {
    if i == 0 {
        return 0
    } else {
        return il[i] + il.offset(i-1)
    }
}

// returns how many snapshots are the goal in the given interval
func (il intervalList) goal(i int) int {
    return int(il[i+1] / il[i])
}

type scheduleList map[string]intervalList

func (schl *scheduleList) String() string {
    a := []string{}
    for sch := range *schl {
        a = append(a, sch)
    }
    return strings.Join(a, ",")
}

/*
  The span of an interval is always the snapshot distance of the next interval.
*/
var schedules = scheduleList{
    "longterm":  {hour * 6, day, week, month, long},
    "shortterm": {minute * 10, hour * 2, day, week, month, long},
    "testing":   {second * 5, second * 20, second * 140, second * 280, long},
    "testing2":  {second * 5, second * 20, second * 40, second * 80, long},
}
