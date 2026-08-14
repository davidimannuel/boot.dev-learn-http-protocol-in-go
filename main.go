package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	var lineChan = make(chan string)
	go func() {
		defer close(lineChan)
		defer f.Close()
		currentline := strings.Builder{}
		buffer := make([]byte, 8)
		for {
			n, err := f.Read(buffer)
			if err != nil && err == io.EOF {
				// TODO: handle error when read in half way
				break
			}
			parts := bytes.Split(buffer[:n], []byte("\n"))
			// core logic is
			// for first element, print with combined latest data of current line if exist with first part
			// for last element, append to current line if exist, later to be combined with first part later
			// other than that always print
			for i := 0; i < len(parts)-1; i++ {
				lineChan <- fmt.Sprintf("read: %s%s", currentline.String(), parts[i])
				currentline.Reset()
			}
			_, err = currentline.Write(parts[len(parts)-1])
			if err != nil {
				panic(err)
			}
		}
	}()

	return lineChan
}

func main() {
	inputFilePath := "messages.txt"
	f, err := os.Open(inputFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for currentline := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", currentline)
	}
}
