// aiffinfo is a command line tool to gather information about aiff/aifc files.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattetti/audio/aiff"
)

const (
	// Height per channel.
	ImgHeight = 400
)

var ImgWidth = 2048
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
		c := aiff.NewDecoder(f, ch)
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
	f.Seek(0, 0)
	info, frames, err := aiff.ReadFrames(f)
	if err != nil {
		log.Fatal(err)
	}

	imgFile, err := os.Create("waveform.png")
	if err != nil {
		log.Fatal(err)
	}
	defer imgFile.Close()

	fmt.Println("sampleRate", info.SampleRate)
	fmt.Println("sampleSize", info.BitsPerSample)
	fmt.Println("numChans", info.NumChannels)
	fmt.Printf("frames: %d\n", len(frames))
	fmt.Println(c)

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

	img := image.NewRGBA(image.Rect(0, 0, ImgWidth, ImgHeight*int(info.NumChannels)))
	if err != nil {
		log.Fatal(err)
	}

	// instead of graphing all points, we only take a proportional sample based on
	// the width of the image
	sampling := len(frames) / ImgWidth
	samplingCounter := 0
	smplBuf := make([][]int, sampling)
	smpl := 0
	for i := 0; i < len(frames); i++ {
		for channel := 0; channel < int(info.NumChannels); channel++ {
			v := frames[i][channel]

			// drawing in the rectable, y=0 is the max, y=height-1 = is the minimun
			// y=height/2 is thw halfway point.
			if v > 0 {
				v = (frames[i][channel] * ImgHeight / 2) / max
				v = ImgHeight/2 - v
			} else {
				v = (abs(frames[i][channel]) * ImgHeight / 2) / max
				v = ImgHeight/2 + v
			}

			if samplingCounter != sampling {
				if channel == 0 {
					smplBuf[samplingCounter] = make([]int, info.NumChannels)
				}
				smplBuf[samplingCounter][channel] = v
				if channel == (info.NumChannels - 1) {
					samplingCounter++
				}
				continue
			}
			v = avg(smplBuf[channel])
			samplingCounter = 0
			smpl++

			// max
			img.Set(smpl, 0, color.RGBA{255, 0, 0, 255})
			// half
			img.Set(smpl, ImgHeight/2, color.RGBA{255, 255, 255, 127})
			// min
			img.Set(smpl, ImgHeight-1, color.RGBA{255, 0, 0, 255})

			img.Set(smpl, v, color.Black)
			// 2nd point to make it thicker
			img.Set(smpl, v+1, color.Black)
		}
	}

	png.Encode(imgFile, img)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func avg(xs []int) int {
	var total int
	for i := 0; i < len(xs); i++ {
		total += xs[i]
	}
	return total / len(xs)
}
