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
    return fmt.Sprintf("\"%s\"", strings.Join(*o, ""))
}

// Opts setter
func (o *Opts) Set(value string) error {
    *o = strings.Split(value, " ")
    return nil
}

// use own struct as "backing store" for parsed flags
type Config struct {
    rsyncPath   string
    rsyncOpts   Opts
    origin      string
    repository  string
    schedule    string
}

func (c *Config) String() string {
    if c.origin != "" {
        return fmt.Sprintf("Repository: %s, Origin: %s", c.repository, c.origin)
    } else {
        return fmt.Sprintf("Repository: %s", c.repository)
    }
}

var cmd string = ""

func usage() {
    fmt.Printf(`usage: %s <command> <options>
Commands:
    run     Periodically create snapshots
    list    List snapshots
    help    Show usage instructions
Use <command> -h to show possible options for <command>.
Examples:
    %s run -origin=fileserver:/export/projects -repository=/snapshots/projects
    %s list -repository=/snapshots/projects
`, myName, myName, myName)
}

func LoadConfig() *Config {
    config := new(Config)
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
                "rsyncOpts",
                "additional options for rsync")
            flags.StringVar(&(config.origin),
                "origin", "/tmp/snaprd_test/",
                "data source")
            flags.StringVar(&(config.repository),
                "repository", "/tmp/snaprd_dest",
                "where to store snapshots")
            flags.StringVar(&(config.schedule),
                "schedule", "longterm",
                "choose a schedule")
            flags.Parse(os.Args[2:])
            log.Println(cmd, config)
            if _, ok := schedules[config.schedule]; ok == false {
                log.Fatalln("no such schedule:", config.schedule)
            }
            return config
        }
    case "list":
        {
            flags := flag.NewFlagSet(cmd, flag.ExitOnError)
            flags.StringVar(&(config.repository),
                "repository", "/tmp/snaprd_dest",
                "where snapshots are located")
            flags.Parse(os.Args[2:])
            fmt.Println(config)
            return config
        }
    case "help", "-h", "--help":
        {
            usage()
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
