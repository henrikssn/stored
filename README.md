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

This will expose a standard RESTful API on port 8080. It uses the URL as key and stores whatever you put in the body. Choosing a data format is up to you.

I.e.

    PUT /foo
    Body: Bar

will store Bar under the foo key and

    GET /foo

will return Bar in the Body of the HTTP response.

Stored supports GET, PUT and DELETE actions.

Connecting more servers
-----------------------

You can at any time connect more servers to the stored cluster. Stored uses consistent hashing
so previous entries will not be affected.

Status
------

Stored is in early development so do not expect this to be production quality.


License
-------
Apache Commons 2.0
http://www.apache.org/licenses/LICENSE-2.0
