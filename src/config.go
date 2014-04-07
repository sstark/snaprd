package main

type Config struct {
    rsyncPath string
    rsyncOpts []string
    srcPath string
    dstPath string
    wrkPath string
}

func LoadConfig() *Config {
    config := new(Config)
    config.rsyncPath = "/usr/bin/rsync"
    config.rsyncOpts = []string{"-a"}
    config.srcPath = "/tmp/snapr_test/"
    config.dstPath = "/tmp/snapr_dest"
    config.wrkPath = "/tmp"
    return config
}
