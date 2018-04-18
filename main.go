package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thiago/lvm/cmd"
)

var name = "lvm"
var version = "dev"
var commit = ""
var date = ""

func main() {
	app := cmd.App(name, fmt.Sprintf("%s - build date %s - commit: %s", version, date, commit))
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
