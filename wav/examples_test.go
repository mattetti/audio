package wav_test

import (
	"fmt"
	"log"
	"os"

	"github.com/mattetti/audio/wav"
)

func ExampleDecoder_Duration() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dur, err := wav.NewDecoder(f).Duration()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s duration: %s\n", f.Name(), dur)
	// Output: fixtures/kick.wav duration: 204.172335ms
}
