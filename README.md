# tribble-tracker v2
==========
Copyright (c) 2017 The LineageOS Project
==========

NOTE: THIS IS NOT YET COMPLETE. IT WILL NOT WORK.

tribble-tracker v1 was a python flask application. We quickly ran into GIL limitations (too many incoming connections),
and CPU problems (computing aggregates is expensive). This project attempts to solve both of these.

## Design
==========
tribble-tracker v2 has a single handler (/api/v1/stats), which allows you to report statistics from a device. All other
files are rendered to HTML immediately after launch, and every 12 hours after that. You can see an example configuration
file to serve these files in configs/nginx.conf.

Incoming statistics are stored in both a global `statistics` collection, and added/updated in an `aggregate` collection.
The `aggregate` collection is later used for rendering, and the `statistics` table is periodically anonymized and
published on github (https://github.com/lineageos/stats).

## Launching
==========

To build, run `go get` and `go build`. The (single) executable will be created as ./tribble-tracker. To launch with the
defaults, just run `tracker`. The following config options are also available:

```
-host: host to run on (defaults to 0.0.0.0)
-port: port to run on (defaults to 8080)
-mongo: mongo config url (format is [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][?options]).
```




