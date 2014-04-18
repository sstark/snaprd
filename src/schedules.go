package main

const minute int64 = 60
const hour int64 = minute*60
const day int64 = hour*24
const week int64 = day*7
const month int64 = week*4 //month == 4 weeks
const future int64 = 9999999999 //the date this program will stop working

/*
  The span of an interval is always the snapshot distance of the next interval.
*/
var schedules = map[string][]int64{
    "longterm":     {hour*6, day, week, month, future},
    "shortterm":    {600, hour*2, day, week, month, future},
    "testing":      {5, 20, 140, 560, future},
}
