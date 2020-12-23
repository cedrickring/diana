package main

import (
	"github.com/cedrickring/diana/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.Run(); err != nil {
		logrus.Fatal(err)
	}
}
