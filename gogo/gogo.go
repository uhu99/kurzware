package main

import (
	"fmt"
	"os"
)

func main() {
	home := os.Getenv("HOME")
	os.Setenv("GOPATH", home + "/Documents/source/go")
	gopath := os.Getenv("GOPATH")
	fmt.Println("GOPATH: " + gopath)

	os.Setenv("GOROOT", "/usr/local/go")
	goroot := os.Getenv("GOROOT")
	fmt.Println("GOROOT: " + goroot)

	path := os.Getenv("PATH")
	os.Setenv("PATH",
		goroot + "/bin:" +
		gopath + "/bin:" +
		path )

	os.Chdir(gopath + "/src")
}