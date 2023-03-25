package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
)

// bufferedRead reads the contents of an io.Reader using a byte buffer of the specified size
func bufferedRead(data io.Reader, size int) int {
	buff := make([]byte, size)
	var n int
	for {
		b, err := data.Read(buff)
		n += b
		if err == io.EOF {
			break
		}
	}
	return n
}

func main() {
	size := flag.Int("b", 256, "The size of the buffer used to read a file's content")
	check := flag.String("c", "", "The checksum to verify against")
	flag.Parse()

	args := os.Args
	if len(args) < 2 {
		fmt.Println("You must provide a path to a file")
		os.Exit(1)
	}
	// read path from CLI args and validate it
	path := args[len(args)-1]
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hash := md5.New()
	// pipe file content to the md5 buffer
	reader := io.TeeReader(file, hash)
	n := bufferedRead(reader, *size)
	// compute checksum
	checksum := hex.EncodeToString(hash.Sum(nil))

	fmt.Printf("MD5 checksum for %s (%d bytes): %s\n", path, n, checksum)
	if *check != "" {
		fmt.Printf("Checksums match: %t\n", *check == checksum)
	}
	os.Exit(0)
}
