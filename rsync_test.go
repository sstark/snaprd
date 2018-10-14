/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
	"os"
	"path/filepath"
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
	var config = config
	config.repository = "testdata"
	config.ReadCache()
	cmd := createRsyncCommand(testSnapshots[1], testSnapshots[0])
	got := cmd.Args
	wanted := []string{"/usr/bin/rsync", "--delete", "-a", "--stats",
		"--link-dest=testdata/.data/1400337531-1400338693-complete",
		"/tmp/snaprd_test/",
		"testdata/.data/1400534523-0-incomplete"}
	if !reflect.DeepEqual(got, wanted) {
		t.Errorf("wanted %v, got %v", wanted, got)
	}
}

func TestFakeRsyncOk(t *testing.T) {
	var testSnapshots = snapshotList{
		{time.Unix(1400337531, 0), time.Unix(1400338693, 0), stateComplete},
		{time.Unix(1400534523, 0), time.Unix(0, 0), stateIncomplete},
	}
	var config = config
	config.repository = "/tmp/snaprd_dest"
	mockRepository()
	config.ReadCache()
	dir, _ := os.Getwd()
	config.RsyncPath = filepath.Join(dir, "fake_rsync")
	config.RsyncOpts.Set("--fake_exit=24")
	_, err := createSnapshot(testSnapshots[0])
	got := err
	if got != nil {
		t.Errorf("createSnapshot() returned an error, but it shouldn't: %v", got)
	}
}

func TestFakeRsyncFail(t *testing.T) {
	var testSnapshots = snapshotList{
		{time.Unix(1400337531, 0), time.Unix(1400338693, 0), stateComplete},
		{time.Unix(1400534523, 0), time.Unix(0, 0), stateIncomplete},
	}
	var config = config
	config.repository = "/tmp/snaprd_dest"
	config.ReadCache()
	dir, _ := os.Getwd()
	config.RsyncPath = filepath.Join(dir, "fake_rsync")
	config.RsyncOpts.Set("--fake_exit=3")
	_, err := createSnapshot(testSnapshots[0])
	got := err
	if got == nil {
		t.Errorf("createSnapshot() succeeded, but it should have failed: %v", got)
	}
}
