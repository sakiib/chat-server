package utils

import "strings"

func ParseUsersList(msg string) []string {
	return strings.Split(msg[strings.Index(msg, "[")+1:len(msg)-1], ",")
}
