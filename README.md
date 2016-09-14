snaprd - backup utility
=======================


Overview
--------

- Continuous creation of snapshots at certain intervals
- Pruning (sieving) snapshots based on fixed schedule, make snapshots more
  scarce the older they get
- Uses rsync to create snapshots
- Every snapshot is a complete copy, using hard links to save disk space
- Designed to run silently in the background
- Repository is designed to be exported via e. g. nfs or smb to enable users to
  do restores of single files or directories

The project homepage is https://gitlab.tuebingen.mpg.de/stark/snaprd


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
