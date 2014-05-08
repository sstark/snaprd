snaprd - backup utility

============================
NOT READY FOR PRODUCTION USE
============================


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

