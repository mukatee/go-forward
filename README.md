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

I shall try to set up a Github releases page. Until then, get the source and compile it. Eh?

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

Notice that the start command uses www.stackoverflow.com and the curl uses the host header of stackoverflow.com.
The initial one just uses that to capture the site address, which is the same for both.
But if you request the site from curl with www.stackoverflow.com host header, you will get back a redirect response
to go to http://stackoverflow.com. 
You can actually try that by adding the www prefix to the host header in curl.
The host header is required because that is what the webserver uses to associate the request to resources.
Browsers do it automatically as does curl but since from their viewpoint the host is localhost:5566, the webserver
will give an error without explicitly setting a different host header.

Anyway, this is just a workable example to illustrate the use.
Most websites tend to use HTTPS these days, so you would also need support to MITM SSL here or something like that in those cases.
Here, curl is just used as an example.
Personally, I have used this for forwarding some custom network traffic to multiple endpoints.
For that, we can specify a mirror address:

```shell
gofwd -dh www.stackoverflow.com -dp 80 -mh localhost -mp 9998 -sp 5566
```

This will not forward all the data coming to localhost:5566 not only to stackoverflow.com:80 but also to localhost:9998.
Sometimes this is handy, for example, I needed to split a datastream from a custom device to two tools that both used it.
In this case, the stream source only supported one target. So there you go, forwarded it to two places.
Learned some Go while at it.

Limitations
-----------

You can only define one mirroring target each (which can be separate for up/downstream). 
It should not be too hard to add more but one has been enough for me.

License
-------

MIT License

