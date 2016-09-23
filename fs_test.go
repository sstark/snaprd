/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
	"log"
	"os"
	"path"
	"syscall"
	"testing"
)

var testDir = os.Getenv("HOME")

func gatherTestData(baseDir string) (data syscall.Statfs_t) {
	err := syscall.Statfs(testDir, &data)
	if err != nil {
		log.Println("could not check free space:", err)
	}
	return

}

func TestCheckFreeSpace(t *testing.T) {
	// First, gather the data
	data := gatherTestData("/")

	var actualFreePerc = 100 * float64(data.Bfree) / float64(data.Blocks)
	var actualFreeGiB = int(uint64(data.Bsize) * data.Bfree / GiB)

	// Now, let's make a quick run of the test
	var result bool
	result = checkFreeSpace(testDir, 0, 0)
	if !result {
		t.Errorf("Short run failure")
	}

	// Successful absolute free space
	result = checkFreeSpace(testDir, 0, actualFreeGiB/2)
	if !result {
		t.Errorf("Error in successful absolute free space test")
	}

	// Successful relative free space
	result = checkFreeSpace(testDir, actualFreePerc/2, 0)
	if !result {
		t.Errorf("Error in successful relative free space test")
	}

	// Successful combined free space
	result = checkFreeSpace(testDir, actualFreePerc/2, actualFreeGiB/2)
	if !result {
		t.Errorf("Error in successful combined free space test")
	}

	// Failed absolute free space
	result = checkFreeSpace(testDir, 0, actualFreeGiB*2)
	if result {
		t.Errorf("Error in failed absolute free space test")
	}

	// Failed relative free space
	result = checkFreeSpace(testDir, actualFreePerc*2, 0)
	if result {
		t.Errorf("Error in failed absolute free space test")
	}

	// Failed combined free space
	result = checkFreeSpace(testDir, actualFreePerc*2, actualFreeGiB*2)
	if result {
		t.Errorf("Error in Failed combined free space test")
	}
}

type dslTestPair struct {
	linkname   string
	target     string
	isDangling bool
}

var dslTestPairs = []dslTestPair{
	dslTestPair{"link1", path.Join(dataSubdir, mockSnapshots[0]), false},
	dslTestPair{"link2", path.Join(dataSubdir, mockSnapshots[1]), false},
	dslTestPair{"link3", "1400337531-1400337532-notexist", false},
	dslTestPair{"link4", path.Join("/absolute", dataSubdir, "1400337531-1400337532-notexist"), false},
	dslTestPair{"link5", "notdatasubdir/1400337531-1400337532-notexist", false},
	dslTestPair{"link6", path.Join(dataSubdir, "/1400337531-1400337532-notexist"), true},
	dslTestPair{"link7", dataSubdir, false},
	dslTestPair{"link8", path.Join(dataSubdir, "/notexist"), true},
}

func TestIsDanglingSymlink(t *testing.T) {
	mockConfig()
	mockRepository()
	defer os.RemoveAll(config.repository)
	for i := range dslTestPairs {
		lname := path.Join(config.repository, dslTestPairs[i].linkname)
		tname := dslTestPairs[i].target
		overwriteSymlink(tname, lname)
		got := isDanglingSymlink(lname)
		wanted := dslTestPairs[i].isDangling
		if got != wanted {
			t.Errorf("%s: got %v, wanted %v", lname, got, wanted)
		}
	}
}

func TestOverwriteSymlink(t *testing.T) {
	mockConfig()
	mockRepository()
	defer os.RemoveAll(config.repository)
	testdir := path.Join(config.repository, "somedir")
	os.Mkdir(testdir, 0777)
	err := overwriteSymlink("irrelevant", testdir)
	if err == nil {
		t.Errorf("%s was overwritten, but it shouldn't", testdir)
	}
	testfile := path.Join(config.repository, "somefile")
	_, _ = os.Create(testfile)
	err = overwriteSymlink("irrelevant", testfile)
	if err == nil {
		t.Errorf("%s was overwritten, but it shouldn't", testfile)
	}
	testlink := path.Join(config.repository, "somelink")
	_ = os.Symlink("irrelevant", testlink)
	err = overwriteSymlink("irrelevant", testlink)
	if err != nil {
		t.Errorf("%s was not overwritten, but it should", testlink)
	}
}
