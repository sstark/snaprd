snaprd - backup utility
=======================


Overview
--------

Snaprd is a program that helps you to make backups of directories to another
file system or server. You can run it on a server that has enough disk space
and let it continuously fetch incremental changes from another server. Snaprd
will make sure that only as many snapshots are created as match your given
schedule or free space restrictions.

- Continuous creation of snapshots at certain intervals
- Pruning (sieving) snapshots based on fixed schedule, make snapshots more
  scarce the older they get
- Pruning snapshots based on free disk space requirements
- Uses rsync to create snapshots
- Every snapshot is a complete copy, using hard links to save disk space
- Designed to run silently in the background
- Repository is designed to be exported via e. g. nfs or smb to enable users to
  do restores of single files or directories
- Tested with small and huge (100TB) backup sources.

The project homepage is https://github.com/sstark/snaprd


Building
--------

Install go either from https://golang.org/ or from your distribution
repository, e. g. "apt-get install golang".

Download the archive, unpack and run `make`. Then copy the binary to a
convenient place or run `make install` to place it in /usr/local/bin

OR

Run `go get github.com/sstark/snaprd`. The binary will be in
`$GOPATH/bin` afterwards.


Installing
----------

Snaprd does not daemonize, logs are printed to the standard output. Choose
whatever you like for starting it at boot: rc.local, SysVinit, upstart,
systemd, supervisord, BSD-init, launchd, daemontools, ...

In case your repository resides in a separate file system you may want to put
some mechanism before startup that makes sure this file system is mounted.


Running
-------

snaprd is supposed to be run on the system where the repository is located
("the backup system"). You will want to make sure ssh works non-interactively
for the connection to the system to backup from.

Basic operation:

```
> snaprd run -schedule=shortterm -repository=/target/dir -origin=someserver:some/dir -noWait
2016/09/14 20:32:29 snaprd 1.0 started with pid 50606
2016/09/14 20:32:29 ### Repository: /target/dir, Origin: someserver:some/dir, Schedule: shortterm
2016/09/14 20:32:29 run: [/usr/bin/rsync --delete -a --link-dest=/target/dir/.data/1473875491-1473875492-complete someserver:some/dir /target/dir/.data/1473877949-0-incomplete]
2016/09/14 20:32:29 finished: 1473877949-1473877950-complete
2016/09/14 20:32:29 wait 9m59.817467794s before next snapshot
[...]
```

```
> snaprd list -repository /tmp/snaprd_dest
### Repository: /tmp/snaprd_dest, Origin: /tmp/snaprd_test2, Schedule: shortterm
### From past, 0/∞
### From 866h0m0s ago, 0/4
### From 194h0m0s ago, 0/7
### From 26h0m0s ago, 2/12
2016-09-14 Wednesday 12:14:31 (1s, 2h0m0s)
2016-09-14 Wednesday 12:19:46 (2s, 2h0m0s)
### From 2h0m0s ago, 5/12
2016-09-14 Wednesday 19:51:07 (1s, 10m0s)
2016-09-14 Wednesday 19:51:21 (1s, 10m0s)
2016-09-14 Wednesday 19:51:26 (1s, 10m0s)
2016-09-14 Wednesday 19:51:31 (1s, 10m0s)
2016-09-14 Wednesday 20:32:29 (1s, 10m0s)
```


Stopping
--------

snaprd will immediately exit when sent the TERM or INT (ctrl-c) signal. If a
backup is running at this time it will be left in incomplete state. (On the next
run it will be reused potentially.)

You can also send the USR1 signal, in which case snaprd will wait until the
current backup has finished, and exit afterwards.

You can find the pid of the running process in the repository directory in the
file `.pid`.


Schedules
---------

There are currently two builtin schedules for snapshots which you can choose
with the -schedule switch to the run command:

  - shortterm: 10m 2h 24h 168h 672h
  - longtterm: 6h 24h 168h 672h

The duration listed define how long a snapshot stays in that interval until it
is either promoted to the next higher interval or deleted.

You can define your own schedules by editing a json-formatted file `/etc/snaprd.schedules` with entries like:

```
{
    "production" : [ {"day":1}, {"week":2}, {"month":1}, {"long":1} ],
    "moreoften" : [ {"hour":6}, {"day":1}, {"week":2}, {"month":1}, {"long":1} ]
}
```

The above 'production' schedule will tell snaprd to make a snapshot every day,
keep one of those every 2 weeks, keep one of those every month. The last entry
("long") should not be omitted and basically means eternity.

As many snapshots are kept as "fit" into the next interval.

The 'moreoften' schedule will do almost the same as 'production', but make
snapshots every 6 hours, thus keeping 4 snapshots per day.

You can verify your schedule by running `snaprd scheds`, and later, when
snapshots have already been created, by `snaprd list`.


Testing
-------

To run regression testing, run `make test`

Debug output can be enabled by setting the environment variable SNAPRD_DEBUG=1

