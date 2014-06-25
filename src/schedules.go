/* See the file "LICENSE.txt" for the full license governing this code. */

// Define snapshot schedules (duration tables) and how to handle them

package main

import (
    "strings"
    "time"
    "encoding/json"
    "io/ioutil"
    "fmt"
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

// Returns how long ago the given interval started
func (il intervalList) offset(i int) time.Duration {
    if i == 0 {
        return 0
    } else {
        return il[i] + il.offset(i-1)
    }
}

// Returns how many snapshots are the goal in the given interval
func (il intervalList) goal(i int) int {
    if i > len(il)-2 {
        panic("this should not happen: highest interval is innumerable!")
    }
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

// List of available snapshot schedules. Defines how often snapshots are made
// or purged. The span of an interval is always the snapshot distance of the
// next interval.
var schedules = scheduleList{
    "longterm":  {hour * 6, day, week, month, long},
    "shortterm": {minute * 10, hour * 2, day, week, month, long},
    "testing":   {second * 5, second * 20, second * 140, second * 280, long},
    "testing2":  {second * 5, second * 20, second * 40, second * 80, long},
}

func (schl scheduleList) AddFromFile(file string) {
    schedFile, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Printf("Error opening schedule file: %v\n", err)
    }
    
    var readData scheduleList

    err = json.Unmarshal(schedFile,&readData)
    if err != nil {
        fmt.Printf("Error parsing data: %v\n", err)
    }
    
    for k,v := range readData {
        schl[k] = v
    }   
}

func (schl scheduleList) List() {
    for name,sched := range schl {
        fmt.Println(name,": ", sched);
    }
}
