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
    verbose     bool
    showAll     bool
}

func (c *Config) String() string {
    if c.origin != "" {
        return fmt.Sprintf("Repository: %s, Origin: %s", c.repository, c.origin)
    } else {
        return fmt.Sprintf("Repository: %s", c.repository)
    }
}

var subcmd string = ""

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
        subcmd = os.Args[1]
    } else {
        log.Fatal("no subcommand given")
    }
    switch subcmd {
    case "run":
        {
            flags := flag.NewFlagSet(subcmd, flag.ExitOnError)
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
            schedulesArr := []string{}
            for sch := range schedules {
                schedulesArr = append(schedulesArr, sch)
            }
            flags.StringVar(&(config.schedule),
                "schedule", "longterm",
                "one of " + strings.Join(schedulesArr, ","))
            flags.Parse(os.Args[2:])
            log.Println(subcmd, config)
            if _, ok := schedules[config.schedule]; ok == false {
                log.Fatalln("no such schedule:", config.schedule)
            }
            return config
        }
    case "list":
        {
            flags := flag.NewFlagSet(subcmd, flag.ExitOnError)
            flags.StringVar(&(config.repository),
                "repository", "/tmp/snaprd_dest",
                "where snapshots are located")
            flags.BoolVar(&(config.verbose),
                "v", false,
                "show more information")
            flags.BoolVar(&(config.showAll),
                "a", false,
                "show all snapshots. Otherwise only complete snapshots are shown")
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
            log.Println("unknown subcommand:", subcmd)
            log.Fatalln("try \"help\"")
        }
    }
    return nil
}
