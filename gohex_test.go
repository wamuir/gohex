package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteByteSlice(t *testing.T) {

	s := "Hello, hexadecimal world!"
	stdin := bytes.NewBufferString(s)
	stdout := bytes.NewBuffer(make([]byte, 0, len(s)))

	err := writeByteSlice(stdin, stdout)
	assert.Nil(t, err)

	exp := "\t" + `0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x68, 0x65, 0x78,` + "\n"
	exp += "\t" + `0x61, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x20, 0x77,` + "\n"
	exp += "\t" + `0x6f, 0x72, 0x6c, 0x64, 0x21,` + "\n"

	assert.Equal(t, []byte(exp), stdout.Bytes())
	return
}
