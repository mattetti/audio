/*
Package aiff is a AIFF/AIFC decoder and encoder.
It extracts the basic information of the file and provide decoded frames for AIFF files.
Other chunks, including AIFC frames can be accessed by using a channel.

Besides the parsing functionality, this package also provides a way for developers to access the duration of an aiff/aifc file.
For an example of how to use the custom parser, look at the aiffinfo CLI tool which uses `NewDecoder` constructor and a channel
to receive chunks.

This package also allows for quick access to the AIFF LPCM raw audio data:

    in, err := os.Open("audiofile.aiff")
    if err != nil {
    	log.Fatal("couldn't open audiofile.aiff %v", err)
    }
    info, frames, err := ReadFrames(in)
    in.Close()

A frame is a slice where each entry is a channel and each value is the sample value.
For instance, a frame in a stereo file will have 2 entries (left and right) and each entry will
have the value of the sample.


Finally, the encoder allows the encoding of LPCM audio data into a valid AIFF file.
Look at the encoder_test.go file for a more complete example.

Currently only AIFF is properly supported, AIFC files will more than likely not be properly processed.

*/
package aiff
