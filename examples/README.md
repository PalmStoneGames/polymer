Polymer examples
=================

Each subfolder contains a specific example case.

Requires [GopherJs](https://github.com/gopherjs/gopherjs) and [Caddy](https://github.com/mholt/caddy). Instructions [below](#prerequisites).

### Quickstart
```
$ cd examples
$ ./build-all.sh
$ caddy
```
Point browser to `http://localhost:2015` and navigate to any of the example directories.

### Build specific example 
```
$ cd examples/example_name
$ gopherjs build
$ cd ..
$ caddy
```
Point browser to `http://localhost:2015/example_name`

### Prerequisites 
GopherJS 
```
$ go get github.com/gopherjs/gopherjs
```
Caddy 
```
$ go get github.com/mholt/caddy
```
