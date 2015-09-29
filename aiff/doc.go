/*
Package aiff is a AIFF/AIFC parser that extracts the basic information and allows developers to customize the parsing experienced.

Besides the parsing functionality, this package also provides a way for developers to access the duration of an aiff/aifc file.
For an example of how to use the custom parser, look at the aiffinfo CLI tool which uses `NewParser` constructor and a channel
to receive chunks.


Currently only AIFF is properly supported, AIFC files will more than likely not be properly processed.

*/
package aiff
