package main

import (
    "bufio"
    "io"
    "log"
    "os"
    "os/exec"
    "path/filepath"
)

func createRsyncCommand(sn *Snapshot, base *Snapshot) *exec.Cmd {
    cmd := exec.Command(config.rsyncPath)
    args := make([]string, 0, 256)
    args = append(args, config.rsyncPath)
    args = append(args, "-a")
    args = append(args, config.rsyncOpts...)
    if base != nil {
        args = append(args, "--link-dest="+base.FullName())
    }
    args = append(args, config.origin, sn.FullName())
    cmd.Args = args
    cmd.Dir = filepath.Join(config.repository, DATA_SUBDIR)
    log.Println("run:", args)
    return cmd
}

func runRsyncCommand(cmd *exec.Cmd) error {
    var err error
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return err
    }
    stdoutReader := bufio.NewReader(stdout)
    stderrReader := bufio.NewReader(stderr)
    err = cmd.Start()
    if err != nil {
        return err
    }
    for {
        str, err := stderrReader.ReadString('\n')
        if err == io.EOF {
            break
        }
        if err != nil && err != io.EOF {
            return err
        }
        log.Print("<rsync stderr> ", str)
    }
    for {
        str, err := stdoutReader.ReadString('\n')
        if err == io.EOF {
            break
        }
        if err != nil && err != io.EOF {
            return err
        }
        log.Print("<rsync stdout> ", str)
    }
    err = cmd.Wait()
    if err != nil {
        return err
    }
    return nil
}

func CreateSnapshot(base *Snapshot) {
    // first snapshot
    if base == nil {
        path := filepath.Join(config.repository, DATA_SUBDIR)
        log.Println("creating destination directory for initial snapshot:", path)
        err := os.MkdirAll(path, 00755)
        if err != nil {
            log.Fatal(err)
        }
    }
    newSn := newIncompleteSnapshot()
    cmd := createRsyncCommand(newSn, base)
    err := runRsyncCommand(cmd)
    if err != nil {
        log.Fatalln("rsync error:", err)
    }
    newSn.transComplete()
    log.Println("finished:", newSn.Name())
}
