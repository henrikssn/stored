Stored
======

A fast available distributed in-memory key-value store written in Go.

Installation
------------

    go install github.com/henrikssn/stored

Example
-------

Start a stored deamon using

    stored

This will expose a standard RESTful API on port 8080. It uses the URL as key and stores whatever you put in the body. Choosing a data format is up to you.

I.e. using curl

    curl -X PUT --data "bar" localhost:8080/default/foo/0

will store the string "bar" in the default namespace using the foo key under the 0 id and

    curl -X GET localhost:8080/default/foo/0

will return bar in the body of the HTTP response. There are plans to support returning all ids under the foo key using /default/foo uri but this is not implemented yet.

Using the store client this can be achieved by

    store put foo bar
    store get foo

Stored supports GET, PUT and DELETE actions.

HTTP Errors
-----------
Stored returns standard HTTP statuses when applicable.

PUT:
HTTP 201 if the key was added
HTTP 200 if the existing key was updated.

GET:
HTTP 200 if the key was found.

DELETE:
HTTP 200 if the key was deleted.

All actions:
HTTP 400 if the request was malformed.
HTTP 500 if another error occured.

Connecting more servers
-----------------------

You can at any time connect more servers to the stored cluster. Stored uses consistent hashing
so previous entries will not be affected.

System description
------------------
Stored uses four parts to set up a cluster.
- Store (The actual store)
- Router (Knows where a key resides)
- Endpoint (Talks HTTP)
- StoreMonkey (Health detection and scaling)

Status
------

Stored is in early development so do not expect this to be production quality. Look at the tests to see what is currently working. Default starts up 1 endpoint, 1 router, 1 store and lets your store data with the HTTP interface for testing purposes. You should use a more carefully designed cluster if this is storing important data.


License
-------
Apache Commons 2.0
http://www.apache.org/licenses/LICENSE-2.0
