// aiffinfo is a command line tool to gather information about aiff/aifc files.
// Note that github.com/llgcode/draw2d is a dependency to run this code and generate the waveform.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/mattetti/audio/aiff"
)

const (
	// Height per channel.
	chanHeight = 400
	ImgWidth   = 2048
)

var pathToParse = flag.String("path", ".", "Where to find aiff files")
var fileToParse = flag.String("file", "", "The wav file to analyze (instead of a path)")
var logChunks = flag.Bool("v", false, "Should the parser log chunks (not SSND)")
var waveformNameFlag = flag.String("waveform", "waveform.png", "the filename of the waveform output")

type point struct {
	X, Y float64
}

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

	fmt.Println("sample Rate", info.SampleRate)
	fmt.Println("sample Size", info.BitsPerSample)
	fmt.Println("number of Channels", info.NumChannels)
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

	imgHeight := chanHeight * int(info.NumChannels)
	img := image.NewRGBA(image.Rect(0, 0, ImgWidth, imgHeight))
	if err != nil {
		log.Fatal(err)
	}
	gc := draw2dimg.NewGraphicContext(img)

	gc.SetLineWidth(1)
	// min (max y == bottom of the graph)
	gc.MoveTo(0, float64(imgHeight-1))
	gc.LineTo(ImgWidth, float64(imgHeight-1))
	gc.SetStrokeColor(color.RGBA{255, 255, 255, 100})
	gc.Stroke()

	for i := 0; i < info.NumChannels; i++ {
		// max for chan
		gc.MoveTo(0, float64(i*chanHeight+1))
		gc.LineTo(ImgWidth, float64(i*chanHeight+1))
		gc.SetStrokeColor(color.RGBA{255, 255, 255, 100})
		gc.Stroke()
		// middle
		gc.MoveTo(0, float64(i*chanHeight+(chanHeight/2)))
		gc.LineTo(ImgWidth, float64(i*chanHeight+(chanHeight/2)))
		gc.SetStrokeColor(color.RGBA{255, 255, 255, 127})
		gc.Stroke()
	}

	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})

	gc.SetLineWidth(2)
	// instead of graphing all points, we only take an averaged sample based on
	// the width of the image
	sampling := len(frames) / ImgWidth
	samplingCounter := make([]int, info.NumChannels)
	smplBuf := make([][]int, info.NumChannels)
	for i := 0; i < info.NumChannels; i++ {
		smplBuf[i] = make([]int, sampling)
	}
	smpl := 0
	// last channel position so we can better render multi channel files
	lastChanPos := make([]*point, info.NumChannels)

	for i := 0; i < len(frames); i++ {
		for channel := 0; channel < int(info.NumChannels); channel++ {
			if i == 0 {
				lastChanPos[channel] = &point{
					X: 0,
					Y: float64((channel * chanHeight) + chanHeight/2),
				}
			}
			lastPos := lastChanPos[channel]
			gc.MoveTo(lastPos.X, lastPos.Y)

			v := frames[i][channel]

			// y=0 is the max, y=height-1 = is the minimun
			// y=height/2 is the halfway point. We need to convert our values
			// to conform.
			if v > 0 {
				v = (v * chanHeight / 2) / max
				// positive number, we need to go towards 0 (max value)
				v = chanHeight/2 - v
			} else {
				v = (abs(v) * chanHeight / 2) / max
				// negative number, we want to go away from 0
				v = chanHeight/2 + v
			}

			// adjust the position for the channel we are on
			v = (channel * chanHeight) + v

			// if we aren't "sampling" this sample, we still gather the values
			// to report an average when we actually do sample the value. (this avoids drawing
			// outliers).
			if samplingCounter[channel] != sampling {
				// set the sample buffer value for this channel at this position
				smplBuf[channel][samplingCounter[channel]] = v
				samplingCounter[channel]++
				continue
			}
			// average the skipped samples to avoid drawing an outliner
			v = avg(smplBuf[channel])
			samplingCounter[channel] = 0

			pos := &point{X: float64(smpl), Y: float64(v)}
			gc.LineTo(pos.X, pos.Y)
			gc.Stroke()
			lastChanPos[channel] = pos
			smpl++
		}
	}

	err = draw2dimg.SaveToPngFile(*waveformNameFlag, img)
	if err != nil {
		log.Fatal(err)
	}

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
