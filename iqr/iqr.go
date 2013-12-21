package main

// http://godoc.org/code.google.com/p/rsc/qr
// https://code.google.com/p/freetype-go/
// http://blog.golang.org/go-imagedraw-package

import (
	"strings"
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"code.google.com/p/rsc/qr"
	"code.google.com/p/freetype-go/freetype"
)

var (
	dpi      = flag.Float64("dpi", 180, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "OCRB.ttf", "filename of the ttf font")
	size     = flag.Float64("size", 8, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
	text1    = flag.String("t1", "http://kurzware.de/q", "first line")
	text2    = flag.String("t2", "130342", "second line")
	textqr   = flag.String("tqr", "http://kurzware.de/q?r=130342", "QR text")
)

func main() {
	flag.Parse()

	text := []*string{ text1, text2} //, text3 }

	// Read the font data.
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	// Initialize the context.
	fg, bg := image.Black, image.White
	ruler := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	if *wonb {
		fg, bg = image.White, image.Black
		ruler = color.RGBA{0x22, 0x22, 0x22, 0xff}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 400, 57))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(font)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)

	// Draw the guidelines.
	for i := 0; i < 10; i++ {
		rgba.Set(0, i, ruler)
		rgba.Set(i, 0, ruler)
		rgba.Set(0, 56-i, ruler)
		rgba.Set(i, 56, ruler)

		rgba.Set(399, i, ruler)
		rgba.Set(399-i, 0, ruler)
		rgba.Set(399, 56-i, ruler)
		rgba.Set(399-i, 56, ruler)
	}

	// Draw the text.
	pt := freetype.Pt(10, 2+int(c.PointToFix32(*size)>>8))
	for _, s := range text {
		_, err = c.DrawString(*s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFix32(*size * *spacing)
	}

	// QR Code einfügen
//	qrc := get_qr("http://kurzware.de/q?r=130342")
	qrc,_,_ := image.Decode(strings.NewReader(get_qr(*textqr)))
	dp := image.Rect(330,0, 400,57)
	sp := image.Pt(4,4)
	draw.Draw(rgba, dp, qrc, sp, draw.Src)

	// Save that RGBA image to disk.
	f, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}

//func get_qr(t string) image.Image {
func get_qr(t string) string {
	var q *qr.Code
	var e error

	q,e = qr.Encode(t, qr.L)
	if ( e != nil ) {
		fmt.Println(e)
	}

	fmt.Println(t)
	fmt.Print("QR Code Size (Pixels) = ")
	fmt.Println(q.Size)
	q.Scale = 2

//	return q.Image()
	return string(q.PNG())
}
