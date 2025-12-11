package strings

func TruncateString(s string, maxLength int) string {
	// Convert the string to a slice of runes to handle multi-byte characters correctly.
	runes := []rune(s)

	if len(runes) <= maxLength {
		return s
	}

	// Return runes back as string
	return string(runes[:maxLength])
}
