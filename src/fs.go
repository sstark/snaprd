/* See the file "LICENSE.txt" for the full license governing this code. */

// Low-level filesystem utilities

package main

import (
    "log"
    "syscall"
)

const GiB = 1024 * 1024 * 1024 // One gibibyte (2^30)

// Function to verify the space constraints specified by the user.
// Return true if all the constraints are satisfied, or in case something unusual happens.
func checkFreeSpace(baseDir string, minPerc float64, minGiB int) bool {
    // This is just to avoid the system call if there is nothing to check
    if minPerc <= 0 && minGiB <= 0 {
        return true
    }

    var stats syscall.Statfs_t
    Debugf("Trying to check free space in %s", baseDir)
    err := syscall.Statfs(baseDir, &stats)
    if err != nil {
        log.Println("could not check free space:", err)
        return true // We cannot return false if there is an error, otherwise we risk deleting more than we should
    }

    sizeBytes := uint64(stats.Bsize) * stats.Blocks
    freeBytes := uint64(stats.Bsize) * stats.Bfree

    Debugf("We have %f GiB, and %f GiB of them are free.", float64(sizeBytes)/GiB, float64(freeBytes)/GiB)

    // The actual check... we fail it we are below either the absolute or the relative value

    if int(freeBytes/GiB) < minGiB || (100*float64(freeBytes)/float64(sizeBytes)) < minPerc {
        return false
    }

    return true
}
