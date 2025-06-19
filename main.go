package main

import (
	"github.com/sirupsen/logrus"
	"gvm/cmd"
	_ "gvm/languages/golang"
	_ "gvm/languages/node"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		logrus.Fatalf("%v", err)
	}
}
