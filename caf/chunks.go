package caf

type stringsChunk struct {
	// The number of information strings in the chunk. Must always be valid.
	numEntries uint32
	// Apple reserves keys that are all lowercase (see Information Entry Keys).
	// Application-defined keys should include at least one uppercase character.
	// For any key that ends with ' date' (that is, the space character followed
	// by the word 'date'—for example, 'recorded date'), the value must be a
	// time-of-day string. See Time Of Day Data Format. Using a '.' (period)
	// character as the first character of a key means that the key-value pair
	// is not to be displayed. This allows you to store private information that
	// should be preserved by other applications but not displayed to a user.
	stringID map[string]string
}
