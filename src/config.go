package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strings"
    "encoding/json"
    "path/filepath"
    "io/ioutil"
)

const (
    myName = "snaprd"
    DATA_SUBDIR = ".data"
)

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
    RsyncPath  string
    RsyncOpts  Opts
    Origin     string
    Repository string
    Schedule   string
    Verbose    bool
    ShowAll    bool
    MaxKeep    int
    NoPurge    bool
}

func (c *Config) String() string {
    if c.Origin != "" {
        return fmt.Sprintf("Repository: %s, Origin: %s", c.Repository, c.Origin)
    } else {
        return fmt.Sprintf("Repository: %s", c.Repository)
    }
}

func (c *Config) WriteCache() error {
    cacheFile := filepath.Join(c.Repository, "."+myName+".settings")
    Debugf("trying to write cache settings to %s", cacheFile)
    jsonConfig, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        log.Println("could not write config:", err)
        return err
    }
    err = ioutil.WriteFile(cacheFile, jsonConfig, 0644)
    return err
}

func (c *Config) ReadCache() error {
    t := new(Config)
    cacheFile := filepath.Join(c.Repository, "."+myName+".settings")
    Debugf("trying to read cache settings from %s", cacheFile)
    b, err := ioutil.ReadFile(filepath.Join(c.Repository, "."+myName+".settings"))
    if err != nil {
        return err
    } else {
        err = json.Unmarshal(b, &t)
        if err != nil {
            return err
        }
        c.RsyncPath = t.RsyncPath
        c.RsyncOpts = t.RsyncOpts
        c.Origin    = t.Origin
        c.Schedule  = t.Schedule
        c.MaxKeep   = t.MaxKeep
        c.NoPurge   = t.NoPurge
    }
    return nil
}

var subcmd string = ""

func usage() {
    fmt.Printf(`usage: %[1]s <command> <options>
Commands:
    run     Periodically create snapshots
    list    List snapshots
    help    Show usage instructions
Use <command> -h to show possible options for <command>.
Examples:
    %[1]s run -origin=fileserver:/export/projects -repository=/snapshots/projects
    %[1]s list -repository=/snapshots/projects
`, myName)
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
            flags.StringVar(&(config.RsyncPath),
                "rsyncPath", "/usr/bin/rsync",
                "path to rsync binary")
            flags.Var(&(config.RsyncOpts),
                "rsyncOpts",
                "additional options for rsync")
            flags.StringVar(&(config.Origin),
                "origin", "/tmp/snaprd_test/",
                "data source")
            flags.StringVar(&(config.Repository),
                "repository", "/tmp/snaprd_dest",
                "where to store snapshots")
            flags.StringVar(&(config.Schedule),
                "schedule", "longterm",
                "one of " + schedules.String())
            flags.IntVar(&(config.MaxKeep),
                "maxKeep", 0,
                "how many snapshots to keep in highest (oldest) interval. Use 0 to keep all")
            flags.BoolVar(&(config.NoPurge),
                "noPurge", false,
                "if set, obsolete snapshots will not be deleted")
            flags.Parse(os.Args[2:])
            log.Println(subcmd, config)
            if _, ok := schedules[config.Schedule]; ok == false {
                log.Fatalln("no such schedule:", config.Schedule)
            }
            err := config.WriteCache()
            if err != nil {
                log.Printf("could not write settings cache file:", err)
            }
            return config
        }
    case "list":
        {
            flags := flag.NewFlagSet(subcmd, flag.ExitOnError)
            flags.StringVar(&(config.Repository),
                "repository", "/tmp/snaprd_dest",
                "where snapshots are located")
            flags.BoolVar(&(config.Verbose),
                "v", false,
                "show more information")
            flags.BoolVar(&(config.ShowAll),
                "a", false,
                "show all snapshots. Otherwise only complete snapshots are shown")
            flags.StringVar(&(config.Schedule),
                "schedule", "longterm",
                "one of " + schedules.String())
            flags.Parse(os.Args[2:])
            fmt.Println(config)
            err := config.ReadCache()
            if err != nil {
                log.Println("error reading cached settings (using defaults):", err)
            }
            Debugf("cached config: %s", config)
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
