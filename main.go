package main

import (
	"flag"
)

var token = flag.String("token", "", "your token")
var icqtoken = flag.String("icqtoken", "", "your icq token")
var icqto string
// var icqto = flag.String(&icqto, "icqto", "", "icq id")

func init(){
	flag.StringVar(&icqto, "icqto", "", "icq id")
}

// var isSandbox = flag.Bool("is_sandbox", true, "is sandbox env")

func main() {
	flag.Parse()
	authICQ(*icqtoken)
	rest()
	// Chase()
}
