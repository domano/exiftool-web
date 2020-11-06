package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

func main() {
	var inBuf bytes.Buffer
	var outBuf bytes.Buffer

	err := runExifListx(&inBuf)

	if err != nil {
		panic(err.Error())
	}
	DecodeXML(&inBuf, &outBuf)

	for {
		b := make([]byte, 40)
		_, err := outBuf.Read(b)
		if err != nil {
			panic(err.Error())
		}
		println(string(b))
	}

}

func runExifListx(w io.Writer) error {
	cmd := exec.Command("exiftool", "-listx")
	cmd.Stdout = w
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("exiftool -listx failed: %w", err)
	}
	return nil
}
