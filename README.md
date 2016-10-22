DoBoot
======
[![Build Status on TravisCI](https://secure.travis-ci.org/tueftler/doboot.png)](http://travis-ci.org/tueftler/doboot)

Docker yields a "start" event when containers start. This, however, does not mean they can be considered up and running. DoBoot fills this gap.


```
+-----------------+             +---------+             +--------------+
|                 |             |         |             |              |
| Docker Daemon   |             | DoBoot  | 3) started  | Application  |
|                 | 1) started! |         | and ready!  | using Docker |
|                 +-------------> *magic* +-------------> Socket, e.g. |
|                 |             | v       |             | Traefik      |
|                 |             | |       |             |              |
+-------------+---+             +-+-------+             +--------------+
              |                   |
      +-------v---+               | 2) ready?
      |           |               |
      | Container <---------------+
      |           |
      +-----------+

`------------------------------------´`--------------------------------´
 /var/run/docker.sock                  /var/run/doboot.sock
```