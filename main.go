package main

import (
	"log"
	"runtime"
)

func main() {
	log.Printf("%s", runtime.Version())
	log.Printf("GOOS=%s", runtime.GOOS)
	log.Printf("GOARCH=%s", runtime.GOARCH)
	//log.Printf("GOARM=%s", runtime.GOARM) // NB there is no GOARM.
}
