/* See the file "LICENSE.txt" for the full license governing this code. */

// Global configuration with disk caching
// Parsing of command line flags

package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
)

const (
    myName      = "snaprd"
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
    repository string
    Schedule   string
    verbose    bool
    showAll    bool
    MaxKeep    int
    NoPurge    bool
    NoWait     bool
}

// WriteCache writes the global configuration to disk as a json file.
func (c *Config) WriteCache() error {
    cacheFile := filepath.Join(c.repository, "."+myName+".settings")
    Debugf("trying to write cached settings to %s", cacheFile)
    jsonConfig, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        log.Println("could not write config:", err)
        return err
    }
    err = ioutil.WriteFile(cacheFile, jsonConfig, 0644)
    return err
}

// ReadCache reads from the json configuration cache and resets assorted global
// configuration values from it.
func (c *Config) ReadCache() error {
    t := new(Config)
    cacheFile := filepath.Join(c.repository, "."+myName+".settings")
    Debugf("trying to read cached settings from %s", cacheFile)
    b, err := ioutil.ReadFile(filepath.Join(c.repository, "."+myName+".settings"))
    if err != nil {
        return err
    } else {
        err = json.Unmarshal(b, &t)
        if err != nil {
            return err
        }
        c.RsyncPath = t.RsyncPath
        c.RsyncOpts = t.RsyncOpts
        c.Origin = t.Origin
        if _, ok := schedules[t.Schedule]; ok == false {
            log.Fatalln("no such schedule:", t.Schedule)
        }
        c.Schedule = t.Schedule
        c.MaxKeep = t.MaxKeep
        c.NoPurge = t.NoPurge
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
            flags.StringVar(&(config.repository),
                "repository", "/tmp/snaprd_dest",
                "where to store snapshots")
            flags.StringVar(&(config.Schedule),
                "schedule", "longterm",
                "one of "+schedules.String())
            flags.IntVar(&(config.MaxKeep),
                "maxKeep", 0,
                "how many snapshots to keep in highest (oldest) interval. Use 0 to keep all")
            flags.BoolVar(&(config.NoPurge),
                "noPurge", false,
                "if set, obsolete snapshots will not be deleted")
            flags.BoolVar(&(config.NoWait),
                "noWait", false,
                "if set, skip the initial waiting time before the first snapshot")
            flags.Parse(os.Args[2:])
            if _, ok := schedules[config.Schedule]; ok == false {
                log.Fatalln("no such schedule:", config.Schedule)
            }
            path := filepath.Join(config.repository, DATA_SUBDIR)
            Debugf("creating repository:", path)
            err := os.MkdirAll(path, 00755)
            if err != nil {
                log.Fatal(err)
            }
            err = config.WriteCache()
            if err != nil {
                log.Printf("could not write settings cache file:", err)
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
            flags.StringVar(&(config.Schedule),
                "schedule", "longterm",
                "one of "+schedules.String())
            flags.Parse(os.Args[2:])
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
