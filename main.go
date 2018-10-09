/* See the file "LICENSE.txt" for the full license governing this code. */

// Functions according to sub commands given on command line
// Main snapshot creation and purging loops

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const initialWait = time.Second * 30

var config *Config
var logger *log.Logger

func debugf(format string, args ...interface{}) {
	if os.Getenv("SNAPRD_DEBUG") == "1" {
		logger.Output(2, "<DEBUG> "+fmt.Sprintf(format, args...))
	}
}

// lastGoodTicker is the clock for the create loop. It takes the last
// created snapshot on its input channel and outputs it on the output channel,
// but only after an appropriate waiting time. To start things off, the first
// lastGood snapshot has to be read from disk.
func lastGoodTicker(in, out chan *snapshot, cl clock) {
	var gap, wait time.Duration
	var sn *snapshot
	sn = lastGoodFromDisk(cl)
	if sn != nil {
		debugf("lastgood from disk: %s\n", sn.String())
	}
	// kick off the loop
	go func() {
		in <- sn
		return
	}()
	for {
		sn := <-in
		if sn != nil {
			gap = cl.Now().Sub(sn.startTime)
			debugf("gap: %s", gap)
			wait = schedules[config.Schedule][0] - gap
			if wait > 0 {
				log.Println("wait", wait, "before next snapshot")
				time.Sleep(wait)
				debugf("Awoken at %s\n", cl.Now())
			}
		}
		out <- sn
	}
}

// subcmdRun is the main, long-running routine and starts off a couple of
// helper goroutines.
func subcmdRun() (ferr error) {
	pl := newPidLocker(filepath.Join(config.repository, ".pid"))
	pl.Lock()
	defer pl.Unlock()
	if !config.NoWait {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("waiting %s before making snapshots\n", initialWait)
		select {
		case <-sigc:
			return errors.New("-> Early exit")
		case <-time.After(initialWait):
		}
	}
	createExit := make(chan bool)
	createExitDone := make(chan error)
	// The obsoleteQueue should not be larger than the absolute number of
	// expected snapshots. However, there is no way (yet) to calculate that
	// number.
	obsoleteQueue := make(chan *snapshot, 10000)
	lastGoodIn := make(chan *snapshot)
	lastGoodOut := make(chan *snapshot)
	// Empty type for the channel: we don't care about what is inside, only
	// about the fact that there is something inside
	freeSpaceCheck := make(chan struct{})

	cl := new(realClock)
	go lastGoodTicker(lastGoodIn, lastGoodOut, cl)

	// Snapshot creation loop
	go func() {
		var lastGood *snapshot
		var createError error
	CREATE_LOOP:
		for {
			select {
			case <-createExit:
				debugf("gracefully exiting snapshot creation goroutine")
				lastGoodOut = nil
				break CREATE_LOOP
			case lastGood = <-lastGoodOut:
				sn, err := createSnapshot(lastGood)
				if err != nil || sn == nil {
					debugf("snapshot creation finally failed (%s), the partial transfer will hopefully be reused", err)
					//createError = err
					//go func() { createExit <- true; return }()
				}
				lastGoodIn <- sn
				debugf("pruning")
				prune(obsoleteQueue, cl)
				// If we purge automatically all the expired snapshots,
				// there's nothing to remove to free space.
				if config.NoPurge {
					debugf("checking space constraints")
					freeSpaceCheck <- struct{}{}
				}
			}
		}
		createExitDone <- createError
	}()
	debugf("started snapshot creation goroutine")

	// Usually the purger gets its input only from prune(). But there could be
	// snapshots left behind from a previously failed snaprd run, so we fill
	// the obsoleteQueue once at the beginning.
	for _, sn := range findDangling(cl) {
		obsoleteQueue <- sn
	}

	// Purger loop
	go func() {
		for {
			if sn := <-obsoleteQueue; !config.NoPurge {
				sn.purge()
			}
		}
	}()
	debugf("started purge goroutine")

	// If we are going to automatically purge all expired snapshots, we
	// needn't even starting the gofunc
	if config.NoPurge {
		// Free space claiming function
		go func() {
			for {
				// Wait until we are ordered to do something
				<-freeSpaceCheck
				// Get all obsolete snapshots
				// This returns a sorted list
				snapshots, err := findSnapshots(cl)
				if err != nil {
					log.Println(err)
					return
				}
				if len(snapshots) < 2 {
					log.Println("less than 2 snapshots found, not pruning")
					return
				}
				obsolete := snapshots.state(stateObsolete, none)
				// We only delete as long as we need *AND* we have something to delete
				for !checkFreeSpace(config.repository, config.MinPercSpace, config.MinGiBSpace) && len(obsolete) > 0 {
					// If there is not enough space, purge the oldest snapshot
					last := len(obsolete) - 1
					obsolete[last].purge()
					// We remove it from the list, it's quicker than recalculating the list.
					obsolete = obsolete[:last]
				}
			}
		}()
	}

	// Global signal handling
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	select {
	case sig := <-sigc:
		debugf("Got signal %s", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("-> Immediate exit")
		case syscall.SIGUSR1:
			log.Println("-> Graceful exit")
			createExit <- true
			ferr = <-createExitDone
		}
	case ferr = <-createExitDone:
	}
	return
}

// subcmdList give the user an overview of what's in the repository.
func subcmdList(cl clock) {
	intervals := schedules[config.Schedule]
	if cl == nil {
		cl = new(realClock)
	}
	snapshots, err := findSnapshots(cl)
	if err != nil {
		log.Println(err)
	}
	for n := len(intervals) - 2; n >= 0; n-- {
		debugf("listing interval %d", n)
		if config.showAll {
			snapshots = snapshots.state(any, none)
		} else {
			snapshots = snapshots.state(stateComplete, none)
		}
		snapshots := snapshots.interval(intervals, n, cl)
		debugf("snapshots in interval %d: %s", n, snapshots)
		if n < len(intervals)-2 {
			fmt.Printf("### From %s ago, %d/%d\n", intervals.offset(n+1), len(snapshots), intervals.goal(n))
		} else {
			if config.MaxKeep == 0 {
				fmt.Printf("### From past, %d/∞\n", len(snapshots))
			} else {
				fmt.Printf("### From past, %d/%d\n", len(snapshots), config.MaxKeep)
			}
		}
		for i, sn := range snapshots {
			stime := sn.startTime.Format("2006-01-02 Monday 15:04:05")
			var dur, dist time.Duration
			if i < len(snapshots)-1 {
				dist = snapshots[i+1].startTime.Sub(sn.startTime)
			}
			if sn.endTime.After(sn.startTime) {
				dur = sn.endTime.Sub(sn.startTime)
			}
			if config.verbose {
				fmt.Printf("%d %s (%s, %s/%s, %s) \"%s\"\n", n, stime, dur, intervals[n], dist, sn.state, sn.Name())
			} else {
				fmt.Printf("%s (%s, %s)\n", stime, dur, intervals[n])
			}
		}
	}
}

func main() {
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	var err error
	if config, err = loadConfig(); err != nil || config == nil {
		log.Fatal(err)
	}
	if config.NoLogDate {
		log.SetFlags(logger.Flags() - log.Ldate - log.Ltime)
		logger.SetFlags(logger.Flags() - log.Ldate - log.Ltime)
	}
	switch subcmd {
	case "run":
		log.Printf("%s %s started with pid %d\n", myName, version, os.Getpid())
		log.Printf("### Repository: %s, Origin: %s, Schedule: %s\n", config.repository, config.Origin, config.Schedule)
		err = subcmdRun()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	case "list":
		fmt.Printf("### Repository: %s, Origin: %s, Schedule: %s\n", config.repository, config.Origin, config.Schedule)
		subcmdList(nil)
	case "scheds":
		schedules.list()
	}
	os.Exit(0)
}
