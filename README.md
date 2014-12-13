Stored
======

A fast distributed in-memory key-value store written in Go.

Installation
------------

    go get github.com/henrikssn/stored

Example
-------

Start a stored deamon using

    stored

You can then use the store command util (github.com/henrikssn/store) to try out the server.

    $ store put testkey testvalue
    2014/12/14 00:15:46 tag:42

    $ store get testkey
    2014/12/14 00:16:24 tag:42 status:OK key:"testkey" value:"testvalue"

Connecting more servers
-----------------------

You can at any time connect more servers to the stored cluster. Stored uses consistent hashing
so previous entries will not be affected.

    $ store -l localhost:8000
    $ store -c localhost:8000 -l localhost:8001
    $ store -c localhost:8000 -l localhost:8002

Status
------

Stored is in early development so do not expect this to be production quality.


License
-------
Apache Commons 2.0
http://www.apache.org/licenses/LICENSE-2.0
