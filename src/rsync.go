/* See the file "LICENSE.txt" for the full license governing this code. */
package main

import (
    "errors"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
)

func createRsyncCommand(sn *Snapshot, base *Snapshot) *exec.Cmd {
    cmd := exec.Command(config.RsyncPath)
    args := make([]string, 0, 256)
    args = append(args, config.RsyncPath)
    args = append(args, "-a")
    args = append(args, config.RsyncOpts...)
    if base != nil {
        args = append(args, "--link-dest="+base.FullName())
    }
    args = append(args, config.Origin, sn.FullName())
    cmd.Args = args
    cmd.Dir = filepath.Join(config.repository, DATA_SUBDIR)
    log.Println("run:", args)
    return cmd
}

func runRsyncCommand(cmd *exec.Cmd) (error, chan error) {
    var err error
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    Debugf("starting rsync command")
    err = cmd.Start()
    if err != nil {
        return err, nil
    }
    done := make(chan error)
    go func() {
        done <- cmd.Wait()
    }()
    return nil, done
}

func CreateSnapshot(base *Snapshot, kill chan bool) error {
    newSn := newIncompleteSnapshot()
    cmd := createRsyncCommand(newSn, base)
    err, done := runRsyncCommand(cmd)
    if err != nil {
        log.Fatalln("could not start rsync command:", err)
    }
    Debugf("rsync started")
    for {
        select {
        case <-kill:
            Debugf("trying to kill rsync")
            err := cmd.Process.Signal(syscall.SIGTERM)
            if err != nil {
                log.Fatal("failed to kill: ", err)
            }
            return errors.New("rsync killed by request")
        case err := <-done:
            Debugf("received something on done channel:", nil)
            if err != nil {
                // At this stage rsync ran, but with errors.
                // Restart in case of
                // - temporary network error
                // - disk full?
                // Detect external signalling?
                return err
            }
            newSn.transComplete()
            log.Println("finished:", newSn.Name())
            return nil
        }
    }
}
