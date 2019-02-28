![](snaprd_logo_sq.png?raw=true)

snaprd - backup utility
=======================

[![Build Status](https://travis-ci.com/sstark/snaprd.svg?branch=master)](https://travis-ci.com/sstark/snaprd)

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

Instead of building the program yourself you can download the latest
linux binary from https://github.com/sstark/snaprd/releases. This is usually
built using the latest Ubuntu LTS release.

Build yourself:

Install go either from https://golang.org/ or from your distribution
repository, e. g. "apt-get install golang".

Download the archive, unpack and run `make`. Then copy the binary to a
convenient place or run `make install` to place it in /usr/local/bin

OR

Run `go get github.com/sstark/snaprd`. The binary will be in
`$GOPATH/bin` afterwards.


Installing
----------

Snaprd does not daemonize, logs are printed to the standard output, also the
stdout and stderr of the rsync command that is being run. Choose whatever you
like for starting it at boot: rc.local, SysVinit, upstart, systemd,
supervisord, BSD-init, launchd, daemontools, ...

In case your repository resides in a separate file system you may want to put
some mechanism before startup that makes sure this file system is mounted.

See below for an example how to run snaprd using systemd.


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

The above command will create a hard-linked copy (see the `--link-dest` option
in *rsync(1)*) via ssh from the directory "some/dir" on "someserver" every 10
minutes. The copy will be written into a directory within the `.data`
sub-directory of the target directory "/target/dir" on the local system. The
directory name for the snapshots consists of the start time and end time of the
rsync run (in unix time) and the state of the snapshot. While rsync is running
the name will be `<start>-0-incomplete`. Only after rsync is done, the
directory will be renamed to `<start>-<end>-complete`.

After each snapshot snaprd will also create user-friendly names as symlinks
into the .data dir, so if you export the snapshot directory read-only, users
should find a resonably convenient way to find their backups.

Next, snaprd will *prune* the existing snapshots. That means it will check if a
snapshot is suitable for being advanced into the next level of the schedule (in
this example that means the "two-hourly" interval) or if it should stay in the
current interval. If the current interval is "full" already, but no snapshot is
suitable for being advanced, snaprd will *obsolete* as many snapshots as needed
to match the schedule.

Marking a snapshot "obsolete" simply means renaming it to
`<start>-<end>-obsolete`. From then on it will not show up anymore in normal
listings and also not be considered as a target for --link-dest. The default
for snaprd is to eventually mark those obsolete snapshots as
`<start>-<end>-purging` and delete them from disk. You can tweak this behaviour
with the "-maxKeep", "-noPurge", "-minGbSpace" and "-minPercSpace" parameters
for snaprd.

To get a full list of options available to the run command, use `snaprd run -h`:

    $ snaprd run -h
    Usage of run:
    -maxKeep int
            how many snapshots to keep in highest (oldest) interval. Use 0 to keep all
    -minGbSpace int
            if set, keep at least x GiB of the snapshots filesystem free
    -minPercSpace float
            if set, keep at least x% of the snapshots filesystem free
    -noLogDate
            if set, does not print date and time in the log output. Useful if output is redirected to syslog
    -noPurge
            if set, obsolete snapshots will not be deleted (minimum space requirements will still be honoured)
    -noWait
            if set, skip the initial waiting time before the first snapshot
    -notify string
            specify an email address to send reports
    -origin string
            data source (default "/tmp/snaprd_test/")
    -r string
            (shorthand for -repository) (default "/tmp/snaprd_dest")
    -repository string
            where to store snapshots (default "/tmp/snaprd_dest")
    -rsyncOpts value
            additional options for rsync
    -rsyncPath string
            path to rsync binary (default "/usr/bin/rsync")
    -schedFile string
            path to external schedules (default "/etc/snaprd.schedules")
    -schedule string
            one of longterm,shortterm (default "longterm")


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

The above list command will output some information about the intervals for the
given schedule and how many snapshots are in them.

Obviously the list command needs to know which schedule was used for creating
the snapshots, but in the above example you can see that no schedule was given
at the command line. This works because snaprd writes all settings that were
used for the last *run* command to the repository as `.snaprd.settings`.


E-Mail Notification
-------------------

If you add the `-notify` option to the run sub-command you will get an email
in case of a problem. Use it like this:

```
> snaprd run -notify root <other options...>
```

If snaprd has a severe problem it will stop execution and send an email to the
specified address, along with the last few lines of log output.

Sending happens through use of the standard mail(1) command, make sure your
system is configured accordingly.


System Prerequisites
--------------------

Obviously you need a file system where you can store enough data to fit the
dataset you are backing up. It is not possible to predict how much space will
be needed for a given schedule and update pattern of data. You should at least
make the snapshot file system such that it can be easily extended if needed.
Starting with a factor of 1.5 to 2 should be sufficient.

If you are using mlocate or a similar mechanism to index your files, make sure
you exclude your snapshot file system from it, e. g. like this:

<pre>
$ cat /etc/updatedb.conf
PRUNE_BIND_MOUNTS="yes"
# PRUNENAMES=".git .bzr .hg .svn"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot <b>/snapshots</b>"
PRUNEFS="NFS nfs nfs4 rpc_pipefs afs binfmt_misc proc smbfs autofs iso9660 ncpfs coda devpts ftpfs devfs devtmpfs fuse.mfs shfs sysfs cifs lustre tmpfs usbfs udf fuse.glusterfs fuse.sshfs curlftpfs ceph fuse.ceph fuse.rozofs ecryptfs fusesmb"
</pre>

If you do not exclude your snapshots you will get enormously big mlocate.db
files with lots of redundant information.



Stopping
--------

snaprd will immediately exit when sent the TERM or INT (ctrl-c) signal. If a
backup is running at this time it will be left in incomplete state. (On the next
run it will be reused potentially.)

You can also send the USR1 signal, in which case snaprd will wait until the
current backup has finished, and exit afterwards.

You can find the pid of the running process in the repository directory in the
file `.pid`.

Signals
-------

snaprd responds to various signals.

  - **INT**, **TERM**: Makes snaprd exit immediately, killing a potentially
    running rsync process using the same signal.
  - **USR1**: While rsync is running, wait until it finished, then exit.
    Otherwise just exit.
  - **USR2**: If snaprd is idle waiting for the next scheduled snapshot,
    sending SIGUSR2 will cancel this waiting time and force an immediate
    snapshot.

Schedules
---------

There are currently two builtin schedules for snapshots which you can choose
with the -schedule switch to the run command:

  - shortterm: 10m 2h 24h 168h 672h
  - longtterm: 6h 24h 168h 672h

The duration listed define how long a snapshot stays in that interval until it
is either promoted to the next higher interval or deleted.

Which schedule you choose is entirely up to you, just make sure the smallest
(first) interval is large enough so the expected runtime of a single rsync
snapshot fits in it with a good margin.

You can define your own schedules by editing a json-formatted file
`/etc/snaprd.schedules` with entries like:

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


Example Unit File for Systemd
-----------------------------

Place in `/etc/systemd/system/snaprd-srv-home.service`

    [Unit]
    Description=snapshots for srv:/homes
    Documentation=https://github.com/sstark/snaprd
    Requires=network.target

    [Service]
    User=root
    StandardOutput=syslog
    ExecStart=/usr/local/bin/snaprd run -noLogDate -notify root -repository=/export/srv-home-snap -origin=srv:/export/homes
    Restart=on-failure
    # for a shared machine you probably want to set the scheduling class:
    IOSchedulingClass=idle

    [Install]
    WantedBy=multi-user.target

Enable with

    sudo systemctl enable snaprd-srv-home && sudo systemctl start snaprd-srv-home

Check logs with 

    journalctl -u snaprd-srv-home


Testing
-------

To run regression testing, run `make test`

Debug output can be enabled by setting the environment variable SNAPRD_DEBUG=1

