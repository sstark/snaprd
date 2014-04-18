package main

import (
    "time"
)

const (
    second = time.Second
    minute = time.Minute
    hour = time.Hour
    day = hour*24
    week = day*7
    month = week*4
    year = day*365
    future = year*100
)

/*
  The span of an interval is always the snapshot distance of the next interval.
*/
var schedules = map[string][]time.Duration{
    "longterm":     {hour*6, day, week, month, future},
    "shortterm":    {second*600, hour*2, day, week, month, future},
    "testing":      {second*5, second*20, second*140, second*560, future},
}
