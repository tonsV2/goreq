# Go Req

Define executable HTTP requests as text files

## Quick start

```sh
make binary && cat requests.http | ./goreq
```

## Install
```sh
make install
```

## Executable script

```sh
echo """#!/usr/local/bin/goreq
GET https://reddit.com HTTP/1.1""" > reddit.http
chmod +x reddit.http
./reddit.http
```

## Development

```sh
make dev
```

## TODO

* Sanitize request
    * Allow missing http version
    * missing newlines
* Pretty print
    * headers
    * body

* Assert response
    * Expect - add expect flag which takes an int representing a http status code and matches it against the last
      request performed - return nonzero exit code in case of mismatch
    * ... Take list of ints?

* Possible do the above in a different app which would read and parse the response (the output of this app)?
