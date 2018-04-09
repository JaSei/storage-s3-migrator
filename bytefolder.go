package main

import (
	"fmt"
)

type byteFolder byte

type byteFolders []byteFolder

func (bf byteFolder) hex() string {
	return fmt.Sprintf("%02X", bf)
}
