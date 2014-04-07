package main

import (
    "os/exec"
    "log"
    "bufio"
    "io"
    //"strings"
)

func createRsyncCommand() *exec.Cmd {
    cmd := exec.Command(config.rsyncPath)
    args := make([]string, 0, 256)
    args = append(args, config.rsyncPath)
    args = append(args, config.rsyncOpts...)
    args = append(args, config.srcPath, config.dstPath)
    cmd.Args = args
    cmd.Dir = config.wrkPath
    log.Println("run:", args)
    return cmd
}

func CreateSnapshot(c chan string) {
    cmd := createRsyncCommand()
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }
    rd := bufio.NewReader(stdout)
    err = cmd.Start()
    if err != nil {
        log.Fatal(err)
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
    var msg string = "Snapshot created"
    c <- msg
}
