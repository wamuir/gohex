/*
Command gohex creates static file imports for Go.  Use gohex [file] for
embedding static assets in Go like you would for C with xxd -i [file].

  go get github.com/wamuir/gohex

This command line tool outputs a Go file containing a byte slice, analogous
to a static C array as might generated by hex dumping a file using xxd with
the -i (include) flag. An intended use of this tool is to allow for the
inclusion of a static file (or multiple static files) within a compiled Go
binary.  This is meant to be a straightforward solution for use simple cases:
no frills, file systems, compression or configuration. gohex is also quite fast
and dumps about 35 percent faster than xxd.

A simple example – converting echoed text to a Go static file

Within a unix shell:

  $ printf "Hello, hexadecimal world!" | gohex
  package main

  var gohex = []byte{
  	0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x68, 0x65, 0x78,
  	0x61, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x20, 0x77,
  	0x6f, 0x72, 0x6c, 0x64, 0x21,
  }

If the output was stored as a .go file, the variable could then be accessed:

  fmt.Printf("%s", string(gohex))

Which would yield:

  Hello, hexadecimal world!

Usage and command line flags

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

Git Repository

Source code is available under an MIT License at
https://github.com/wamuir/gohex.
*/
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode"
)

const hextable = `0123456789abcdef`

var (
	c = flag.Int("c", 10, "number of columns to format per line")
	h = flag.Bool("h", false, "print this summary")
	i = flag.Int("i", 1, "number of tabs to indent the byte slice")
	p = flag.String("p", "main", "name for Go package, or empty for none")
	s = flag.Bool("s", false, "output byte slice without declarations")
	v = flag.String("v", "gohex", "name for Go variable of the byte slice")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of gohex:\n")
	fmt.Fprintf(os.Stderr, "\tgohex [flags] [infile [outfile]]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

// declareGoPkg writes Go package declaration
// Example:  package main
func declareGoPkg(w io.Writer) {

	var declaration = make([]byte, 10+len(*p))
	_ = copy(declaration[0:8], []byte("package "))
	_ = copy(declaration[8:8+len(*p)], *p)
	_ = copy(declaration[8+len(*p):], []byte("\n\n"))

	w.Write(declaration)
}

// openGoVar writes variable declaration and left bracket
// Example:  var gohex = []byte{
func openGoVar(w io.Writer) {

	var (
		j      rune
		k      int
		left   []byte = []byte("var ")
		center []byte = []byte(*v)
		right  []byte = []byte(" = []byte{")
		tab    []byte = []byte("\t")
	)

	declaration := make([]byte, 4+len(center)+10)

	_ = copy(declaration[0:len(left)], left[:])

	// First char of identifier must be a letter (including _)
	j = rune(center[0])
	if !unicode.IsLetter(j) && j != '_' {
		center = append([]byte("_"), center...)
	}

	// All identifier chars must be letters (including _) or digits
	for k = 0; k < len(center); k++ {
		j = rune(center[k])
		if unicode.IsLetter(j) || unicode.IsDigit(j) {
			declaration[len(left)+k] = center[k]
		} else {
			declaration[len(left)+k] = '_'
		}
	}
	_ = copy(declaration[len(left)+len(center):], right[:])

	w.Write(bytes.Repeat(tab, *i-1))
	w.Write(declaration)
	w.Write([]byte("\n"))
}

// closeGoVar writes a right bracket to close variable declaration
// Example:  }
func closeGoVar(w io.Writer) {

	var tab []byte = []byte("\t")
	w.Write(bytes.Repeat(tab, *i-1))
	w.Write([]byte("}\n"))
}

// writeByteSlice writes a byte slice from data provided to the reader
// Example:
//	0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x68, 0x65, 0x78,
//      0x61, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x20, 0x77,
func writeByteSlice(r io.Reader, w io.Writer) error {

	var (
		b   byte
		buf []byte = make([]byte, *c)
		err error
		hex []byte = make([]byte, 6)
		j   int
		n   int
		tab []byte = []byte("\t")
	)

	for {
		n, err = io.ReadFull(r, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}

		w.Write(bytes.Repeat(tab, *i))

		for j = 1; j <= n; j++ {

			b = buf[j-1 : j][0]

			hex[0] = '0'
			hex[1] = 'x'
			hex[2] = hextable[b>>4]
			hex[3] = hextable[b&0x0f]
			hex[4] = ','
			hex[5] = ' '

			if j == n {
				hex[5] = '\n'
			}

			w.Write(hex)
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
	}
}

func main() {

	var (
		err     error
		fname   string
		infile  *os.File
		outfile *os.File
		reader  *bufio.Reader
		writer  *bufio.Writer
	)

	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if *h {
		flag.Usage()
		os.Exit(1)
	}

	if *c < 1 {
		err = errors.New("invalid number of columns (min. 1)")
		fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
		os.Exit(1)
	}

	if *i < 1 {
		err = errors.New("invalid indentation (min. 1)")
		fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
		os.Exit(1)
	}

	if *v == "" {
		err = errors.New("invalid variable name")
		fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
		os.Exit(1)
	}

	switch len(args) {

	case 0:
		infile = os.Stdin
		defer infile.Close()

		outfile = os.Stdout
		defer outfile.Close()

	case 1:
		fname = args[0]
		infile, err = os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
			os.Exit(1)
		}
		defer infile.Close()

		outfile = os.Stdout
		defer outfile.Close()

	case 2:
		fname = args[0]
		infile, err = os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
			os.Exit(1)
		}
		defer infile.Close()

		fname = args[1]
		outfile, err = os.Create(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
			os.Exit(1)
		}
		defer outfile.Close()

	default:
		err = errors.New("invalid number of arguments")
		fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}

	reader = bufio.NewReader(infile)

	writer = bufio.NewWriter(outfile)
	defer writer.Flush()

	if !*s && *p != "" {
		declareGoPkg(writer)
	}

	if !*s {
		openGoVar(writer)
	}

	err = writeByteSlice(reader, writer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gohex: %s\n", err.Error())
		os.Exit(1)
	}

	if !*s {
		closeGoVar(writer)
	}
}
