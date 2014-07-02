/* See the file "LICENSE.txt" for the full license governing this code. */

package main

import (
    "testing"
)

const testDir = "/home"

func gatherTestData(baseDir s) (data syscall.Statfs_t) {
    err := syscall.Statfs(testDir, &data)
    if err != nil {
        log.Println("could not check free space:", err)
    }

}

func TestCheckFreeSapce(t *testing.T) {
    // First, gather the data
    data := gatherTestData("/")

    var float64 actualFreePerc = float64(data.Bfree) / float64(data.Bytes)
    var int actualFreeGiB = int(freeBytes / GiB)

    // Now, let's make a quick run of the test
    result := checkFreeSpace(testDir, 0, 0)
    if !result {
        t.Errorf("Short run failure")
    }

    // Successful absolute free space
    result := checkFreeSpace(testDir, 0, actualFreeGiB/2)
    if result {
        t.Errorf("Error in successful absolute free space test")
    }

    // Successful relative free space
    result := checkFreeSpace(testDir, actualFreePerc/2, 0)
    if result {
        t.Errorf("Error in successful absolute free space test")
    }

    // Successful combined free space
    result := checkFreeSpace(testDir, actualFreePerc/2, actualFreeGiB/2)
    if result {
        t.Errorf("Error in successful combined free space test")
    }

    // Failed absolute free space
    result := checkFreeSpace(testDir, 0, actualFreeGiB*2)
    if result {
        t.Errorf("Error in failed absolute free space test")
    }

    // Failed relative free space
    result := checkFreeSpace(testDir, actualFreePerc*2, 0)
    if result {
        t.Errorf("Error in failed absolute free space test")
    }

    // Failed combined free space
    result := checkFreeSpace(testDir, actualFreePerc*2, actualFreeGiB*2)
    if result {
        t.Errorf("Error in Failed combined free space test")
    }
}
