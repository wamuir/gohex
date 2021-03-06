gohex
=====

[![GoDoc Reference](https://godoc.org/github.com/wamuir/gohex?status.svg)](http://godoc.org/github.com/wamuir/gohex)
[![Build Status](https://travis-ci.org/wamuir/gohex.svg?branch=master)](https://travis-ci.org/wamuir/gohex)
[![Go Report Card](https://goreportcard.com/badge/github.com/wamuir/gohex)](https://goreportcard.com/report/github.com/wamuir/gohex)

# Description

Command gohex creates static file imports for Go.  Use `gohex [file]` for
embedding static assets in Go like you would for C with `xxd -i [file]`.

# Installation

This package can be installed with the `go get` command:

    go get github.com/wamuir/gohex

This command line tool outputs a Go file containing a byte slice, analogous
to a static C array as might generated by hex dumping a file using xxd with
the -i (include) flag. An intended use of this tool is to allow for the
inclusion of a static file (or multiple static files) within a compiled Go
binary.  This is meant to be a straightforward solution for use simple cases:
no frills, file systems, compression or configuration. gohex is also quite fast
and dumps about 35 percent faster than xxd.

# Usage and command line flags

```
Usage of gohex:
        gohex [flags] [infile [outfile]]
Flags:
  -c int
        number of columns to format per line (default 10)
  -h    print this summary
  -i int
        number of tabs to indent the byte slice (default 1)
  -p string
        name for Go package, or empty for none (default "main")
  -s    output byte slice without declarations
  -v string
        name for Go variable of the byte slice (default "gohex")
```

# Example

A simple example – converting echoed text to a Go static file

Within a unix shell:

```shell
$ printf "Hello, hexadecimal world!" | gohex
package main

var gohex = []byte{
      0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x68, 0x65, 0x78,
      0x61, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x20, 0x77,
      0x6f, 0x72, 0x6c, 0x64, 0x21,
}
```

If the output was stored as a .go file, the variable could then be accessed:

    fmt.Printf("%s", string(gohex))

Which would yield:

    Hello, hexadecimal world!
