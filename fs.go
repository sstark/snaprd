/* See the file "LICENSE.txt" for the full license governing this code. */

// Low-level filesystem utilities

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
)

// GiB is exactly one gibibyte (2^30)
const GiB = 1024 * 1024 * 1024

// checkFreeSpace verifies the space constraints specified by the user. Return
// true if all the constraints are satisfied, or in case something unusual
// happens.
func checkFreeSpace(baseDir string, minPerc float64, minGiB int) bool {
	// This is just to avoid the system call if there is nothing to check
	if minPerc <= 0 && minGiB <= 0 {
		return true
	}

	var stats syscall.Statfs_t
	debugf("Trying to check free space in %s", baseDir)
	err := syscall.Statfs(baseDir, &stats)
	if err != nil {
		log.Println("could not check free space:", err)
		// We cannot return false if there is an error, otherwise we risk
		// deleting more than we should
		return true
	}

	sizeBytes := uint64(stats.Bsize) * stats.Blocks
	freeBytes := uint64(stats.Bsize) * stats.Bfree

	debugf("We have %f GiB, and %f GiB of them are free.", float64(sizeBytes)/GiB, float64(freeBytes)/GiB)

	// The actual check... we fail it we are below either the absolute or the
	// relative value

	if int(freeBytes/GiB) < minGiB || (100*float64(freeBytes)/float64(sizeBytes)) < minPerc {
		return false
	}

	return true
}

// updateSymlinks creates user-friendly symlinks to all complete snapshots. It
// also removes symlinks to snapshots that have been purged.
func updateSymlinks() {
	entries, err := ioutil.ReadDir(config.repository)
	if err != nil {
		log.Println("could not read repository directory", config.repository)
		return
	}
	for _, f := range entries {
		pathName := path.Join(config.repository, f.Name())
		if isDanglingSymlink(pathName) {
			debugf("symlink %s is dangling, remove", pathName)
			err := os.Remove(pathName)
			if err != nil {
				log.Println("could not remove link", pathName)
			}
		}
	}
	cl := new(realClock)
	snapshots, err := findSnapshots(cl)
	if err != nil {
		log.Println("could not list snapshots")
		return
	}
	for _, s := range snapshots.state(stateComplete, none) {
		target := path.Join(dataSubdir, s.Name())
		stime := s.startTime.Format("Monday_2006-01-02_15.04.05")
		linkname := path.Join(config.repository, stime)
		overwriteSymlink(target, linkname)
	}
	return
}

// isDanglingSymlink returns true only if linkname is a relative symlink
// pointing to a non-existing path in dataSubdir.
func isDanglingSymlink(linkname string) bool {
	target, err := os.Readlink(linkname)
	if err != nil {
		return false
	}
	//debugf("%s: %v", linkname, err)
	if path.IsAbs(target) {
		return false
	}
	pe := strings.Split(target, "/")
	if len(pe) == 0 || pe[0] != dataSubdir {
		return false
	}
	basedir := path.Dir(linkname)
	_, err = os.Stat(path.Join(basedir, target))
	if err != nil && os.IsNotExist(err) {
		return true
	}
	return false
}

// overwriteSymlink creates a symbolic link from linkname to target. It will
// overwrite an already existing link under linkname, but not if it finds a
// regular file or directory (or anything else which is not a symlink) under
// that name.
func overwriteSymlink(target, linkname string) (err error) {
	fi, err := os.Lstat(linkname)
	if err != nil {
		// link does not exist or can not be read. Ignore.
		//debugf("%v", err)
	}
	if fi != nil {
		// link exists
		if fi.Mode()&os.ModeSymlink != 0 {
			// link is indeed a symlink
			ltarget, lerr := os.Readlink(linkname)
			if lerr != nil {
				debugf("could not read %s: %v", linkname, lerr)
			}
			// short cut if the link is already pointing to the desired target
			if ltarget == target {
				return
			}
			debugf("symlink needs removal: %s != %s", target, ltarget)
			err = os.Remove(linkname)
			if err != nil {
				// link can not be removed
				return
			}
		} else {
			err = fmt.Errorf("won't overwrite %s, it is not a symlink (%v)", linkname, fi.Mode())
			return
		}
	}
	err = os.Symlink(target, linkname)
	if err == nil {
		//debugf("symlink %s -> %s", linkname, target)
	}
	return
}
