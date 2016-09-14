snaprd - backup utility
=======================


Overview
--------

- continuous creation of snapshots at certain intervals
- pruning (sieving) snapshots based on fixed schedule, make
  snapshots more scarce the older they get
- uses rsync to create snapshots
- every snapshots is a complete copy, using hard links to
  save disk space
- designed to run silently in the background
- repository is designed to be exported via e. g. nfs or smb
  to enable users to do restores of single files or directories


Building
--------

Download the archive, unpack and run `make`. Then copy the binary to a
convenient place.

OR

Run `go get gitlab.tuebingen.mpg.de/stark/snaprd`. The binary will be in
`$GOPATH/bin` afterwards.


Testing
-------

To run regression testing, run `make test`
