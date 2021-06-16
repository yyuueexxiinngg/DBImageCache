package utils

import "strings"

func SplitJavID(javID string) (string, string) {
	index := strings.LastIndex(javID, "-")

	return javID[:index], javID[index+1:]
}
