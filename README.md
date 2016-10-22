DoBoot
======
[![Build Status on TravisCI](https://secure.travis-ci.org/tueftler/doboot.png)](http://travis-ci.org/tueftler/doboot)

Docker yields a "start" event when containers start. This, however, does not mean they can be considered up and running. DoBoot fills this gap.


```
+-----------------+             +---------+             +--------------+
|                 |             |         |             |              |
| Docker Daemon   |             | DoBoot  | 3) started  | Application  |
|                 | 1) started! |         | and ready!  | using Docker |
|                 +-------------> [*] >>> +-------------> Socket, e.g. |
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

How it works
------------
When a "start" event is received on the Docker socket, *DoBoot* inspects the container and checks for `boot` label. If present, *DoBoot* runs this command in the container and waits for it to return. The "start" event is then passed on to the DoBoot socket for other applications to consume.

In the other direction, *DoBoot* simply passes all communication on to the Docker socket and replies with what it receives. This means it will work with any application currently using Docker's API to drive its business logic. 

Using DoBoot
------------
The only prerequisite to using *DoBoot* semantics is to add the label and the boot script. To set the label, either use the *LABEL* instruction inside the Dockerfile or pass `-l boot=/boot.sh` on the docker command line.

```Dockerfile
FROM debian:jessie

LABEL boot CMD /boot.sh

HEALTHCHECK CMD /health.sh

CMD ...
```

The boot script may be something as simple as `sleep 5`, but you're encouraged to write more sophisticated checks to determine whether your service is actually up and running.

Contributing
------------

To contribute to DoGet, use the :octocat: way - fork, hack, and submit a pull request! If you're unsure where to start, look out for [issues](https://github.com/tueftler/doboot/issues) labeled with **help wanted**.

**Enjoy!**
