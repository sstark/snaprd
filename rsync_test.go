/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
	"reflect"
	"testing"
	"time"
)

func TestCreateRsyncCommand(t *testing.T) {
	var testSnapshots = snapshotList{
		{time.Unix(1400337531, 0), time.Unix(1400338693, 0), stateComplete},
		{time.Unix(1400534523, 0), time.Unix(0, 0), stateIncomplete},
	}
	// shadow global config
	config = config
	config.repository = "testdata"
	config.ReadCache()
	cmd := createRsyncCommand(testSnapshots[1], testSnapshots[0])
	got := cmd.Args
	wanted := []string{"/usr/bin/rsync", "--delete", "-a",
		"--link-dest=testdata/.data/1400337531-1400338693-complete",
		"/tmp/snaprd_test/",
		"testdata/.data/1400534523-0-incomplete"}
	if !reflect.DeepEqual(got, wanted) {
		t.Errorf("wanted %v, got %v", wanted, got)
	}
}
