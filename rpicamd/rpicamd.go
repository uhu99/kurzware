package main

import (
	"log"
	"flag"
	"net/http"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	dir  = flag.String("dir", ".", "Working Directory")
	port = flag.String("port", "8008", "Server Port")
)

func raspistill(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	fmt.Println("raspistill")
	fmt.Println(values)

	var opt []string
	opt = values["options"]
	if opt != nil {
		opt = strings.Fields(values["options"][0])
		opt = append(opt, "-e", "jpg")
		opt = append(opt, "-o", *dir + "/image.jpg")
	} else {
		opt = []string{"-e", "jpg", "-o", *dir + "/image.jpg"}
	}
	cmd := exec.Command("/opt/vc/bin/raspistill", opt...)

	cmd.Stderr = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Status: %v\n", err)

	http.ServeFile(w, r, *dir + "/image.jpg")
}

func raspiyuv(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	fmt.Println("raspiyuv")
	fmt.Println(values)

	var opt []string
	opt = values["options"]
	if opt != nil {
		opt = strings.Fields(values["options"][0])
		opt = append(opt, "-o", *dir + "/image.yuv")
	} else {
		opt = []string{"-o", *dir + "/image.yuv"}
	}
	cmd := exec.Command("/opt/vc/bin/raspiyuv", opt...)

	cmd.Stderr = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Status: %v\n", err)

	http.ServeFile(w, r, *dir + "/image.yuv")
}

func raspivid(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	fmt.Println("raspivid")
	fmt.Println(values)

	var opt []string
	opt = values["options"]
	if opt != nil {
		opt = strings.Fields(values["options"][0])
		opt = append(opt, "-o", *dir + "/video.h264")
	} else {
		opt = []string{"-o", *dir + "/video.h264"}
	}
	cmd := exec.Command("/opt/vc/bin/raspivid", opt...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Status: %v\n", err)

	http.ServeFile(w, r, *dir + "/video.h264")
}

func main() {
	flag.Parse()
	//port := ":8008"
	//dir := "."

	fmt.Println("ListenAndServe " + *port)
	fmt.Println("FileServerDir  " + *dir)

	http.Handle("/", http.FileServer(http.Dir(*dir)))
	http.HandleFunc("/raspistill", raspistill)
	http.HandleFunc("/raspiyuv", raspiyuv)
	http.HandleFunc("/raspivid.h264", raspivid)
 
	// To serve a directory on disk (/tmp) under an alternate URL
	// path (/tmpfiles/), use StripPrefix to modify the request
	// URL's path before the FileServer sees it:
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}

