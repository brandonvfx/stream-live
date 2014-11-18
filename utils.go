package main

import "strings"

func TruncString(str string, length int) string {
	if len(str) > length {
		return strings.TrimSpace(str[:length]) + "..."
	} else {
		return strings.TrimSpace(str)
	}
}
