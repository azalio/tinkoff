package main

import (
	"flag"
)

var token = flag.String("token", "", "your token")

// var isSandbox = flag.Bool("is_sandbox", true, "is sandbox env")

func main() {
	flag.Parse()
	rest()
	// Chase()
}
