/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
    "strconv"
    "testing"
    "time"
)

var sdate int64 = 1400268377
var edate int64 = 1400268387

func TestNewSnapshot(t *testing.T) {
    out := strconv.FormatInt(sdate, 10)+"-"+strconv.FormatInt(edate, 10)+" Complete"
    sn := newSnapshot(time.Unix(sdate, 0), time.Unix(edate, 0), STATE_COMPLETE)
    if s := sn.String(); s != out {
        t.Errorf("sn.String() = %v, want %v", s, out)
    }
}
