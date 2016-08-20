package presenters

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/mattetti/audio"
)

// CSV writes the content of the buffer in a CSV file.
func CSV(buf *audio.PCMBuffer, path string, format audio.DataFormat) error {
	csvf, err := os.Create(path)
	if err != nil {
		return err
	}
	defer csvf.Close()
	csvw := csv.NewWriter(csvf)
	samples := buf.AsInts()
	row := make([]string, buf.Format.NumChannels)

	for i := 0; i < buf.Format.NumChannels; i++ {
		row[i] = fmt.Sprintf("Channel %d", i+1)
	}
	if err := csvw.Write(row); err != nil {
		return fmt.Errorf("error writing header to csv: %s", err)
	}

	for i := 0; i < len(samples); i++ {
		for j := 0; j < buf.Format.NumChannels; j++ {
			row[j] = strconv.Itoa(samples[i])
			if i >= len(samples) {
				break
			}
		}
		if err := csvw.Write(row); err != nil {
			return fmt.Errorf("error writing record to csv: %s", err)
		}
	}
	csvw.Flush()
	return csvw.Error()
}
