package main

import (
    "time"
    "strings"
)

const (
    second = time.Second
    minute = time.Minute
    hour   = time.Hour
    day    = hour * 24
    week   = day * 7
    month  = week * 4
    year   = day * 365
    future = year * 100
)

type intervalList []time.Duration
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
var schedules = scheduleList {
    "longterm":  {hour * 6, day, week, month, future},
    "shortterm": {minute * 10, hour * 2, day, week, month, future},
    "testing":   {second * 5, second * 20, second * 140, second * 280, future},
    "testing2":  {second * 5, second * 20, second * 140, second * 280},
}
