/* See the file "LICENSE.txt" for the full license governing this code. */

// Define snapshot schedules (duration tables) and how to handle them

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
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

// offset returns how long ago the given interval started
func (il intervalList) offset(i int) time.Duration {
	if i == 0 {
		return 0
	}
	return il[i] + il.offset(i-1)
}

// goal returns how many snapshots are the goal in the given interval
func (il intervalList) goal(i int) int {
	if i > len(il)-2 {
		panic("this should not happen: highest interval is innumerable!")
	}
	return int(il[i+1] / il[i])
}

type scheduleList map[string]intervalList

type jsonInterval []map[string]time.Duration

func (schl *scheduleList) String() string {
	a := []string{}
	for sch := range *schl {
		a = append(a, sch)
	}
	return strings.Join(a, ",")
}

// schedules is a list of available snapshot schedules. Defines how often
// snapshots are made or purged. The span of an interval is always the snapshot
// distance of the next interval.
var schedules = scheduleList{
	"longterm":  {hour * 6, day, week, month, long},
	"shortterm": {minute * 10, hour * 2, day, week, month, long},
}

// addFromFile adds an external JSON file to the list of available scheds
func (schl scheduleList) addFromFile(file string) error {
	// If we are using the default file name, and it doesn't exist, no problem, just return
	if _, err := os.Stat(file); os.IsNotExist(err) && file == defaultSchedFileName {
		return nil
	}
	schedFile, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Error opening schedule file: %v", err)
	}
	var readData map[string]jsonInterval
	err = json.Unmarshal(schedFile, &readData)
	if err != nil {
		return fmt.Errorf("Error parsing schedule file: %v", err)
	}
	for k, v := range readData {
		schl[k] = v.intervalList()
	}
	return nil
}

// list prints the stored schedules in the list
func (schl scheduleList) list() {
	var sKeys []string
	for k := range schl {
		sKeys = append(sKeys, k)
	}
	sort.Strings(sKeys)
	for _, name := range sKeys {
		fmt.Printf("%s: %s\n", name, schl[name])
	}
}

// Transform a JSON formatted intervalList like this:
// [
//   { "day" : 1, "hour" : 12 },
//   { "week" : 2 },
//   { "month" : 1, "week" : 2}
//   { "long" : 1}
// ]
// and it makes it equivalent to
// { 1*day + 12*hour, 2*week, 1*month + 2*week, long }

func (json jsonInterval) intervalList() intervalList {
	il := make(intervalList, len(json))
	for i, interval := range json {
		var duration time.Duration
	Loop:
		for k, v := range interval {
			switch k {
			case "s", "second":
				duration += v * second
			case "m", "minute":
				duration += v * minute
			case "h", "hour":
				duration += v * hour
			case "d", "day":
				duration += v * day
			case "w", "week":
				duration += v * week
			case "M", "month":
				duration += v * month
			case "y", "year":
				duration += v * year
			case "l", "long":
				duration = long
				break Loop
			}
		}
		il[i] = duration
	}
	return il
}
