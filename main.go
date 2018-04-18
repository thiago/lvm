package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thiago/lvm/cmd"
)

var name = "lvm"
var version = "dev"
var gitSHA = ""
var buildDate = ""

func main() {
	app := cmd.App(name, fmt.Sprintf("%s - build date %s - commit: %s", version, buildDate, gitSHA))
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
