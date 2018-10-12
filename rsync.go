/* See the file "LICENSE.txt" for the full license governing this code. */

// Handle creation and cancellation of os level process to create snapshots

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// createRsyncCommand returns an exec.Command structure that, when executed,
// creates a snapshot using rsync. Takes an optional (non-nil) base to be used
// with rsyncs --link-dest feature.
func createRsyncCommand(sn *snapshot, base *snapshot) *exec.Cmd {
	cmd := exec.Command(config.RsyncPath)
	args := make([]string, 0, 256)
	args = append(args, config.RsyncPath)
	args = append(args, "--delete")
	args = append(args, "-a")
	args = append(args, config.RsyncOpts...)
	if base != nil {
		args = append(args, "--link-dest="+base.FullName())
	}
	args = append(args, config.Origin, sn.FullName())
	cmd.Args = args
	cmd.Dir = filepath.Join(config.repository, dataSubdir)
	log.Println("run:", args)
	return cmd
}

// runRsyncCommand executes the given command. On sucessful startup return an
// error channel the caller can receive a return status from.
func runRsyncCommand(cmd *exec.Cmd) (chan error, error) {
	var err error
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	debugf("starting rsync command")
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	done := make(chan error)
	go func() {
		time.Sleep(time.Second)
		done <- cmd.Wait()
		return
	}()
	return done, nil
}

// createSnapshot starts a potentially long running rsync command and returns a
// Snapshot pointer on success.
// For non-zero return values of rsync potentially restart the process if the
// error was presumably volatile.
func createSnapshot(base *snapshot) (*snapshot, error) {
	cl := new(realClock)

	newSn := lastReusableFromDisk(cl)

	if newSn == nil {
		newSn = newIncompleteSnapshot(cl)
	} else {
		newSn.transIncomplete(cl)
	}
	cmd := createRsyncCommand(newSn, base)
	done, err := runRsyncCommand(cmd)
	if err != nil {
		log.Println("could not start rsync command:", err)
		return nil, err
	}
	debugf("rsync started")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-sigc:
			debugf("trying to kill rsync with signal %v", sig)
			err := cmd.Process.Signal(sig)
			if err != nil {
				log.Fatal("failed to kill: ", err)
			}
			return nil, errors.New("rsync killed by request")
		case err := <-done:
			debugf("received something on done channel: %v", err)
			if err != nil {
				// At this stage rsync ran, but with errors.
				// Restart in case of
				// - temporary network error
				// - disk full?
				// Detect external signalling?

				failed := true

				// First, get the error code
				if exiterr, ok := err.(*exec.ExitError); ok { // The return code != 0)
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok { // Finally get the actual status code
						debugf("The error code we got is: %v", status.ExitStatus())
						// status now holds the actual return code

						// Magic number: means some files couldn't be
						// copied because they vanished, so nothing
						// critical. See man rsync
						if status.ExitStatus() == 24 {
							debugf("Some files failed to copy because they were deleted in the meantime, but nothing critical... going on...")
							failed = false
						}
					}
				}
				if failed {
					return nil, fmt.Errorf("rsync failed: %s", err)
				}
			}
			newSn.transComplete(cl)
			log.Println("finished:", newSn.Name())
			return newSn, nil
		}
	}
}
