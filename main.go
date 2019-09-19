package main

import (
	"github.com/lossanarch/mcbackup/cmd"
)

type config struct {
	savesPath  string
	targetPath string
}

var minecraftDir string

func main() {
	cmd.Execute()
}
