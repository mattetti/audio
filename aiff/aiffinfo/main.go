// aiffinfo is a command line tool to gather information about aiff/aifc files.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattetti/audio/aiff"
)

var pathToParse = flag.String("path", ".", "Where to find aiff files")
var fileToParse = flag.String("file", "", "The wav file to analyze (instead of a path)")
var logChunks = flag.Bool("v", false, "Should the parser log chunks (not SSND)")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: \n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *fileToParse != "" {
		analyze(*fileToParse)
		return
	}
	if err := filepath.Walk(*pathToParse, walkFn); err != nil {
		log.Fatal(err)
	}
}

func walkFn(path string, fi os.FileInfo, err error) (e error) {
	if err != nil {
		log.Fatal(err)
	}
	if fi.IsDir() {
		filepath.Walk(path, walkFolder)
		return
	}
	if (!strings.HasSuffix(fi.Name(), ".aif") && !strings.HasSuffix(fi.Name(), ".aiff")) || fi.IsDir() {
		return
	}
	analyze(path)
	return nil
}

func walkFolder(path string, fi os.FileInfo, err error) (e error) {
	if (!strings.HasSuffix(fi.Name(), ".aif") && !strings.HasSuffix(fi.Name(), ".aiff")) || fi.IsDir() {
		return
	}
	analyze(path)
	return nil
}

func analyze(path string) {
	fmt.Println(path)
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if *logChunks {
		ch := make(chan *aiff.Chunk)
		c := aiff.NewParser(f, ch)
		go func() {
			if err := c.Parse(); err != nil {
				panic(err)
			}
		}()

		for chunk := range ch {
			id := string(chunk.ID[:])
			fmt.Println(id, chunk.Size)
			if id != "SSND" {
				buf := make([]byte, chunk.Size)
				chunk.ReadBE(buf)
				fmt.Print(hex.Dump(buf))
			}
			chunk.Done()
		}
		return
	}
	c := aiff.New(f)
	if err := c.Parse(); err != nil {
		log.Fatalf("Can't parse the headers of %s - %s\n", path, err)
	}
	fmt.Println(c)
}
