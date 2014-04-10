package main

import (
    "flag"
    "os"
    "log"
    "fmt"
    "strings"
)

const myName = "snaprd"

type Opts []string

// Opts getter
func (o *Opts) String() string {
    return "\"" + strings.Join(*o, " ") + "\""
}

// Opts setter
func (o *Opts) Set(value string) error {
    fmt.Printf("%s\n", value)
    *o = strings.Split(value, " ")
    return nil
}

// use own struct as "backing store" for parsed flags
type Config struct {
    rsyncPath string
    rsyncOpts Opts
    origin string
    repository string
}

var cmd string = ""

func LoadConfig() *Config {
    config := new(Config)
    config.rsyncOpts = []string{"-a"}
    if len(os.Args) > 1 {
        cmd = os.Args[1]
    } else {
        log.Fatal("no subcommand given")
    }
    switch cmd {
    case "run":
        {
            flags := flag.NewFlagSet(cmd, flag.ExitOnError)
            flags.StringVar(&(config.rsyncPath),
                "rsyncPath", "/usr/bin/rsync",
                "path to rsync binary")
            flags.Var(&(config.rsyncOpts),
                "rsyncOpts", // default value set above
                "additional options for rsync")
            flags.StringVar(&(config.origin),
                "origin", "/tmp/snaprd_test/",
                "data source")
            flags.StringVar(&(config.repository),
                "repository", "/tmp/snaprd_dest",
                "where to store snapshots")
            flags.Parse(os.Args[2:])
            log.Println(cmd, config)
            return config
        }
    case "list":
        {
            flags := flag.NewFlagSet(cmd, flag.ExitOnError)
            flags.StringVar(&(config.repository),
                "repository", "/tmp/snaprd_dest",
                "where snapshots are located")
            flags.Parse(os.Args[2:])
            log.Println(cmd, config)
            return config
        }
    case "help", "-h", "--help":
        {
            fmt.Println("usage info...")
            os.Exit(0)
        }
    default:
        {
            log.Println("unknown subcommand:", cmd)
            log.Fatalln("try \"help\"")
        }
    }
    return nil
}
