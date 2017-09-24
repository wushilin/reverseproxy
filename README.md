# Usage is simple:

$ go get github.com/wushilin/reverseproxy

After that, reverseproxy commandline is $GOPATH/bin/reverseproxy

Usage of reverseproxy:
  -config string
    	Default config json (map url to backend) (default "rp.json")
  -listen string
    	Host and Port to listen (default "0.0.0.0:80")
  -rewrite_host
    	Rewrite host header to backend
  -verbose
    	Verbose logging (default true)


example:
$ reverseproxy -config rp.json -listen 0.0.0.0:80 -rewrite_host=false -verbose=true

a simple rp.json is prefix => destination format:
{
  "/api":"http://localhost:8080/api",
  "/anotherapi":"http://remote.server.com:80/anotherapi"
}

If not specified in rp.json, command line is also accepted.

$ reverseproxy /api:http://localhost:88080/api /anotherapi:http://remote.server.com:80/anotherapi

Note default rewrite_host is false, verbose is true, listen is 0.0.0.0:80 config is rp.json


