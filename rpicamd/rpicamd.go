package main

import (
	"log"
	"net/http"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		opt = append(opt, "-o", "image.jpg")
	} else {
		opt = []string{"-e", "jpg", "-o", "image.jpg"}
	}
	cmd := exec.Command("/opt/vc/bin/raspistill", opt...)

	cmd.Stderr = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Status: %v\n", err)

	http.ServeFile(w, r, "image.jpg")
}

func raspiyuv(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	fmt.Println("raspiyuv")
	fmt.Println(values)

	var opt []string
	opt = values["options"]
	if opt != nil {
		opt = strings.Fields(values["options"][0])
		opt = append(opt, "-o", "image.yuv")
	} else {
		opt = []string{"-o", "image.yuv"}
	}
	cmd := exec.Command("/opt/vc/bin/raspiyuv", opt...)

	cmd.Stderr = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Status: %v\n", err)

	http.ServeFile(w, r, "image.yuv")
}

func raspivid(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	fmt.Println("raspivid")
	fmt.Println(values)

	var opt []string
	opt = values["options"]
	if opt != nil {
		opt = strings.Fields(values["options"][0])
		opt = append(opt, "-o", "video.h264")
	} else {
		opt = []string{"-o", "video.h264"}
	}
	cmd := exec.Command("/opt/vc/bin/raspivid", opt...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Status: %v\n", err)

	http.ServeFile(w, r, "video.h264")
}

func main() {
	port := ":8008"
	dir := "."

	fmt.Println("ListenAndServe " + port)
	fmt.Println("FileServerDir  " + dir)

	http.Handle("/", http.FileServer(http.Dir(dir)))
	http.HandleFunc("/raspistill", raspistill)
	http.HandleFunc("/raspiyuv", raspiyuv)
	http.HandleFunc("/raspivid", raspivid)
 
	// To serve a directory on disk (/tmp) under an alternate URL
	// path (/tmpfiles/), use StripPrefix to modify the request
	// URL's path before the FileServer sees it:
	log.Fatal(http.ListenAndServe(port, nil))
}

