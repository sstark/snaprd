package main

import (
    "os/exec"
    "log"
    "bufio"
    "io"
    "path/filepath"
    "os"
)

func createRsyncCommand(sn *Snapshot, base *Snapshot) *exec.Cmd {
    cmd := exec.Command(config.rsyncPath)
    args := make([]string, 0, 256)
    args = append(args, config.rsyncPath)
    args = append(args, config.rsyncOpts...)
    if base != nil {
        args = append(args, "--link-dest=" + filepath.Join(config.dstPath, base.Name()))
    }
    args = append(args, config.srcPath, filepath.Join(config.dstPath, sn.Name()))
    cmd.Args = args
    cmd.Dir = config.wrkPath
    log.Println("run:", args)
    return cmd
}

func CreateSnapshot(c chan error, base *Snapshot) {
    // first snapshot
    if base == nil {
        log.Println("creating destination directory for initial snapshot:", config.dstPath)
        err := os.MkdirAll(config.dstPath, 00755)
        if err != nil {
            log.Fatal(err)
        }
    }
    newSn := newIncompleteSnapshot()
    cmd := createRsyncCommand(newSn, base)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        c <- err
        return
    }
    rd := bufio.NewReader(stdout)
    err = cmd.Start()
    if err != nil {
        c <- err
        return
    }
    for {
        str, err := rd.ReadString('\n')
        if err == io.EOF {
            break
        }
        if err != nil && err != io.EOF {
            log.Println("Read Error: ", err)
            return
        }
        log.Print(str)
    }
    err = cmd.Wait()
    if err != nil {
        c <- err
        return
    }
    newSn.transComplete()
    c <- nil
}
