package main

// import (
// 	"log"
// )

func stringInSlice(a string, list []string) bool {
	// log.Println(a, list)
	for _, b := range list {
		// log.Println(b, list)
		if b == a {
			return true
		}
		if b+".ME" == a {
			return true
		}
	}
	return false
}
