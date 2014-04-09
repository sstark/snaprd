package main

type Config struct {
    rsyncPath string
    rsyncOpts []string
    srcPath string
    dstPath string
    wrkPath string
}

const myName = "snaprd"

func LoadConfig() *Config {
    config := new(Config)
    config.rsyncPath = "/usr/bin/rsync"
    config.rsyncOpts = []string{"-a"}
    config.srcPath = "/tmp/snaprd_test/"
    config.dstPath = "/tmp/snaprd_dest"
    config.wrkPath = "/tmp"
    return config
}
