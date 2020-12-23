package util

import "os"

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" { // unix
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
