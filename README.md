# Libucl Library for Go

go-libucl is a [libucl](https://github.com/vstakhov/libucl) library for
[Go](http://golang.org). Rather than re-implement libucl in Go, this library
uses cgo to bind directly to libucl. This allows the libucl project to be
the central source of knowledge. This project works on Mac OS X, Linux, and
Windows.

**Warning:** This library is still under development and API compatibility
is not guaranteed. Additionally, it is not feature complete yet, though
it is certainly usable for real purposes (we do!).

## Installation

Because we vendor the source of libucl, you can go ahead and get it directly.
We'll keep up to date with libucl. The package name is `libucl`.

```
$ go get github.com/mitchellh/go-libucl
```

Documentation is available on GoDoc: http://godoc.org/github.com/mitchellh/go-libucl

### Compiling Libucl

Libucl should compile easily and cleanly on POSIX systems.

On Windows, msys should be used. msys-regex needs to be compiled.
