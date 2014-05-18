/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
    "testing"
)

var testSchedules = scheduleList{
    "testSched1": {second, second * 2, second * 4, second * 8, second * 16, long},
}

type offsetTestPair struct {
    i       int
    seconds int64
}

func TestScheduleOffset(t *testing.T) {
    var tests = []offsetTestPair{
        {0, 0},
        {1, 2},
        {2, 6},
        {3, 14},
        {4, 30},
        {5, 3153600030},
    }
    for _, pair := range tests {
        v := int64(testSchedules["testSched1"].offset(pair.i).Seconds())
        if v != pair.seconds {
            t.Errorf("offset(%v) got %v, expected %v", pair.i, v, pair.seconds)
        }
    }
}

type goalTestPair struct {
    i    int
    goal int
}

func TestScheduleGoal(t *testing.T) {
    var tests = []goalTestPair{
        {0, 2},
        {1, 2},
        {2, 2},
        {3, 2},
        {4, 197100000},
    }
    for _, pair := range tests {
        v := testSchedules["testSched1"].goal(pair.i)
        if v != pair.goal {
            t.Errorf("goal(%v) got %v, expected %v", pair.i, v, pair.goal)
        }
    }
}
