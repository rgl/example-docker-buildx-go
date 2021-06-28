package main

import (
	"log"
	"runtime"
)

var (
	TARGETPLATFORM string // NB this is set by the linker with -X.
)

func main() {
	log.SetFlags(0)

	log.Printf("%s", runtime.Version())
	log.Printf("TARGETPLATFORM=%s", TARGETPLATFORM)
	log.Printf("GOOS=%s", runtime.GOOS)
	log.Printf("GOARCH=%s", runtime.GOARCH)
	//log.Printf("GOARM=%s", runtime.GOARM) // NB there is no GOARM.
}
