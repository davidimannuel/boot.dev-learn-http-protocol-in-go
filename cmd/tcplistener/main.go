package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
)

// source could be file, could be tcp connection, etc
func getLinesChannel(source io.ReadCloser) <-chan string {
	var lineChan = make(chan string)
	go func() {
		defer close(lineChan)
		defer source.Close()
		currentline := strings.Builder{}
		buffer := make([]byte, 8)
		for {
			n, err := source.Read(buffer)
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
				lineChan <- fmt.Sprintf("%s%s", currentline.String(), parts[i])
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
	listener, err := net.Listen("tcp", ":42069")
	panicIfErr(err)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		panicIfErr(err)
		fmt.Println("connection accepted")
		for currentline := range getLinesChannel(conn) {
			fmt.Print(currentline)
		}
	}

}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
