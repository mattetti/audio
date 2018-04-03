package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mattetti/audio/caf"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("You need to pass the path of the file to analyze.")
	}

	path := os.Args[1]
	fmt.Println(path)
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open the passed path - %v", err)
	}
	defer f.Close()

	d := caf.NewDecoder(f)
	if err = d.ReadInfo(); err != nil {
		log.Fatalf("Failed to read information - %v", err)
	}
	/*
		var chk *chunk.Reader
		for err == nil {
			chk, err = d.NextChunk()
			if err == nil {
				fmt.Println(string(chk.ID[:]))
				chk.Done()
			}
		}
	*/
	fmt.Println(d)
}
