package fs

import (
	"strconv"
	"strings"
)

// for some parsing stuff
func Unicode(str string) string {
	r, _ := strconv.ParseInt(strings.TrimPrefix(str, "\\U"), 16, 32)
	return string(r)
}