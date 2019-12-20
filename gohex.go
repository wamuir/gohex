package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode"
)

const (
	usage    = `Usage: gohex [infile [outfile]]`
	hextable = `0123456789abcdef`
)

// Command line flags
var (
	goPkgName  *string // Flag: pkg; Default: main
	hexcolumns *int    // Flag: cols; Default: 10
	indent     *int    // Flag: indent; Default 1
	sliceOnly  *bool   // Dlag: strip; Default: false
)

// declareGoPkg writes Go package declaration
// Example:  package main
func declareGoPkg(w io.Writer) {

	var declaration = make([]byte, 10+len(*goPkgName))
	_ = copy(declaration[0:8], []byte("package "))
	_ = copy(declaration[8:8+len(*goPkgName)], *goPkgName)
	_ = copy(declaration[8+len(*goPkgName):], []byte("\n\n"))

	w.Write(declaration)
}

// openGoVar writes variable declaration and left bracket
// Example:  var gohex = []byte{
func openGoVar(w io.Writer, goslice string) {

	var (
		i      int
		j      rune
		left   []byte = []byte("var ")
		center []byte = []byte(goslice)
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
	for i = 0; i < len(center); i++ {
		j = rune(center[i])
		if unicode.IsLetter(j) || unicode.IsDigit(j) {
			declaration[len(left)+i] = center[i]
		} else {
			declaration[len(left)+i] = '_'
		}
	}
	_ = copy(declaration[len(left)+len(center):], right[:])

	w.Write(bytes.Repeat(tab, *indent-1))
	w.Write(declaration)
	w.Write([]byte("\n"))

}

// closeGoVar writes a right bracket to close variable declaration
// Example:  }
func closeGoVar(w io.Writer) {

	var tab []byte = []byte("\t")
	w.Write(bytes.Repeat(tab, *indent-1))
	w.Write([]byte("}\n"))
}

// writeByteSlice writes a byte slice from data provided to the reader
// Example:
//	0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x68, 0x65, 0x78,
//      0x61, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x20, 0x77,
func writeByteSlice(r io.Reader, w io.Writer) error {

	var (
		b   byte
		buf []byte = make([]byte, *hexcolumns)
		err error
		hex []byte = make([]byte, 6)
		i   int
		n   int
		tab []byte = []byte("\t")
	)

	for {
		n, err = io.ReadFull(r, buf)
		if err != nil {
			return err
		}

		w.Write(bytes.Repeat(tab, *indent))

		for i = 1; i <= n; i++ {

			b = buf[i-1 : i][0]

			hex[0] = '0'
			hex[1] = 'x'
			hex[2] = hextable[b>>4]
			hex[3] = hextable[b&0x0f]
			hex[4] = ','
			hex[5] = ' '

			if i == n {
				hex[5] = '\n'
			}

			w.Write(hex)
		}
	}
}

func main() {

	var (
		args    []string
		err     error
		fname   string
		goslice string
		infile  *os.File
		outfile *os.File
		reader  *bufio.Reader
		writer  *bufio.Writer
	)

	goPkgName = flag.String(
		"pkg",
		"main",
		"name for Go package, or blank for none",
	)

	hexcolumns = flag.Int(
		"cols",
		10,
		"number of formatted columns in byte slice",
	)

	indent = flag.Int(
		"indent",
		1,
		"number of tabs to indent byte slice",
	)

	sliceOnly = flag.Bool(
		"strip",
		false,
		"return the byte slice only with no declarations",
	)

	flag.Parse()
	args = flag.Args()

	if *hexcolumns < 1 {
		fmt.Printf("gohex: cols must be at least 1\n")
		os.Exit(1)
	}

	if *indent < 1 {
		fmt.Printf("gohex: indent must be at least 1\n")
		os.Exit(1)
	}

	switch len(args) {

	case 0:
		goslice = "gohex"
		infile = os.Stdin
		defer infile.Close()

		outfile = os.Stdout
		defer outfile.Close()

	case 1:
		fname = args[0]
		goslice = fname
		infile, err = os.Open(fname)
		if err != nil {
			fmt.Printf("gohex: %s\n", err.Error())
			os.Exit(1)
		}
		defer infile.Close()

		outfile = os.Stdout
		defer outfile.Close()

	case 2:
		fname = args[0]
		goslice = fname
		infile, err = os.Open(fname)
		if err != nil {
			fmt.Printf("gohex: %s\n", err.Error())
			os.Exit(1)
		}
		defer infile.Close()

		fname = args[1]
		outfile, err = os.Create(fname)
		if err != nil {
			fmt.Printf("gohex: %s\n", err.Error())
			os.Exit(1)
		}
		defer outfile.Close()

	default:
		fmt.Printf("%s\n", usage)
		os.Exit(1)
	}

	reader = bufio.NewReader(infile)

	writer = bufio.NewWriter(outfile)
	defer writer.Flush()

	if !*sliceOnly && *goPkgName != "" {
		declareGoPkg(writer)
	}

	if !*sliceOnly {
		openGoVar(writer, goslice)
	}

	err = writeByteSlice(reader, writer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		fmt.Printf("gohex: %s\n", err.Error())
		os.Exit(1)
	}

	if !*sliceOnly {
		closeGoVar(writer)
	}

}
