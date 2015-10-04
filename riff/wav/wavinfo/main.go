// wavinfo is a command line tool extracting metadata information from a wav file.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/mattetti/audio/riff"
	"github.com/mattetti/audio/riff/wav"
)

const (
	// Height per channel.
	ImgHeight = 200
)

var (
	inputFlag   = flag.String("path", "../fixtures/kick.wav", "path to the file to parse")
	imgNameFlag = flag.String("img", "waveform.png", "name of the image to generate")
)

func main() {
	flag.Parse()

	f, err := os.Open(*inputFlag)
	if err != nil {
		log.Fatal(err)
	}

	imgFile, err := os.Create(*imgNameFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer imgFile.Close()

	d := wav.NewDecoder(f)
	ch := make(chan *riff.Chunk)

	go func() {
		if err := d.Parse(ch); err != nil {
			log.Fatal(err)
		}
	}()

	var frames [][]int
	for chunk := range ch {
		if chunk.ID == riff.DataFormatID {
			frames, err = d.DecodeRawPCM(chunk)
			if err != nil {
				chunk.Done()
				break
			}
		}
		// without this, the goroutines will deadlock
		chunk.Done()
	}

	if err != nil {
		fmt.Println("something went wrong decoding the PCM data", err)
		os.Exit(1)
	}

	info, err := d.Info()
	if err != nil {
		fmt.Println("something went wrong fetching the file's info", err)
		os.Exit(1)
	}
	fmt.Println(info)
	fmt.Println(len(frames), "audio frames")

	img := image.NewRGBA(image.Rect(0, 0, len(frames), ImgHeight*int(info.NumChannels)))
	if err != nil {
		log.Fatal(err)
	}

	max := 0
	for _, f := range frames {
		for _, v := range f {
			if v > max {
				max = v
			} else if v*-1 > max {
				max = v * -1
			}
		}
	}

	for i := 0; i < len(frames); i++ {
		for channel := 0; channel < int(info.NumChannels); channel++ {
			y := (frames[i][channel]*ImgHeight)/max + ImgHeight*channel
			//if y > 0 {
			//fmt.Println(frames[i][channel], y)
			//}
			img.Set(i, y, color.Black)
			img.Set(i, y+1, color.Black)
		}
	}

	png.Encode(imgFile, img)
}
