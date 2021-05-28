package boolconv

import "strconv"

func IsTrue(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func IsFalse(s string) bool {
	return !IsTrue(s)
}
