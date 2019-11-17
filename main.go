package main

import (
	"flag"
	"fmt"
	"os"
	"usdn/funcs"
	"usdn/utils"
)

// 帮助内容
func usage() {
	fmt.Fprintf(os.Stderr, `ULB Agent
Usage: ./ulbctl [-hl] [-c filename] [-t pid]

Options:
`)
	flag.PrintDefaults()
}

func main() {

	cfg := flag.String("c", "cfg.json", "configuration file")
	flag.Usage = usage

	flag.Parse()
	utils.ParseConfig(*cfg)

	utils.Init()

	funcs.Run()

}
