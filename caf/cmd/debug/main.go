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

	decoder := caf.New(f)
	if err := decoder.Parse(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(decoder)
}
