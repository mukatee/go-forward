Go-Forward
==========

A command line tool for forwarding TCP sockets to one or two destinations.
It is written in Go, or as some refer to it, in Golang.
This was actually my first program in Go, so there you go. 
Do proceed to tell me everything that is wrong with it.

The following examples use the name "gofwd" for the executable.
You could rename it whatever you like of course..

Download
--------

I shall try to set up a Github releases page. Until then, get the source and compile it.

Should be something along the lines of:

```shell
go get -u github.com/mukatee/go-forward
go install github.com/mukatee/go-forward
```

Then find the exe/bin files in your GOPATH/bin dir.

Command line options
--------------------

You can access this help txt by running the command without arguments:

```shell
Usage: gofwd [options]
 Options:
  -bufs int
    	Size of read/write buffering. (default 1024)
  -ddf string
    	If defined, will write downstream data to this file.
  -dh string
    	Destination host to forward incoming connections. Required.
  -dp int
    	Destination port to forward incoming connections. Required.
  -duf string
    	If defined, will write upstream data to this file.
  -logc
    	If defined, write debug log info to console.
  -logf string
    	If defined, will write debug log info to this file.
  -mdh string
    	Mirror host to forward incoming connection downstream traffic. Optional.
  -mdp int
    	Mirror port to forward incoming connection downstream data. Optional. Required if downstream mirror host is defined.
  -muh string
    	Mirror host to forward incoming connection upstream traffic. Optional.
  -mup int
    	Mirror port to forward incoming connection upstream data. Optional. Required if upstream mirror host is defined.
  -sp int
    	Source port for incoming connections. Required.
```

Example uses from command line
------------------------------

Example 1: Forward local port 5566 to stackoverflow.com port 80:

```shell
gofwd -dh www.stackoverflow.com -dp 80 -sp 5566
```

Then to request the data via the forwarder:

```shell
curl localhost:5566 --header 'Host: stackoverflow.com'
```

The host header is required because the webserver uses it to associate the request to resources.
Browsers add it automatically, as does curl, but since curl thinks is is requesting the host at localhost:5566, 
it will add a host header for "localhost:5566".
The webserver will then give an error without explicitly setting a different host header as is done here.

This is just a workable example to illustrate the use.
For HTTPS (used by most sites), a different approach would be needed in any case.
Curl is just used as a (hopefully) more understandable example.
Personally, I have mainly used this for forwarding some custom network traffic to multiple endpoints.
For that, we can specify a mirror address:

Example 2: Forward with mirroring:
```shell
gofwd -dh www.stackoverflow.com -dp 80 -mh localhost -mp 9998 -sp 5566
```

This will now forward all the data coming to localhost:5566, not only to stackoverflow.com:80, but also to localhost:9998.
Sometimes this is handy.
For example, I needed to split a datastream from a custom device to two tools that both used it.
In this case, the stream source only supported one target. 
So there you go, used this to forward it to two places.
Probably there are a bunch of command line tools, that I don't know of, to do this already.
Well, at least I learned some Go while at it.

Few more:

Example 3: Forward with saving upstream data to file:
```shell
gofwd -dh www.stackoverflow.com -dp 80 -mh localhost -mp 9998 -sp 5566 -duf uplink.txt
```

Example 4: Forward with saving upstream and downstream data to files:
```shell
gofwd -dh www.stackoverflow.com -dp 80 -mh localhost -mp 9998 -sp 5566 -duf uplink.txt -daf downlink.txt
```

Limitations
-----------

You can only define one mirroring target each (which can be separate for up/downstream). 
It should not be too hard to add more but one has been enough for me.

License
-------

MIT License

