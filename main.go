package main

import (
	"fmt"
	"os"
	"github.com/terrascope/geometry"
	"image"
	"image/png"
	"source.cloud.google.com/wald-1526877012527/cloud_wms/rastreader"
	"time"
)

func main() {
	bbox := geometry.BBox(0, 0, 10, 10)
	if bbox.Area() > 25000000000 {
		fmt.Println("Too big")
		return
	}

	var d time.Time

	out, err := rastreader.GenerateModisTile(256, 256, bbox, d)
	fmt.Println(err)
	fmt.Println(out[0:100])
	fmt.Println(len(out))

	img := &image.Gray{out, 256, image.Rect(0, 0, 256, 256)}

	f, _ := os.Create("./yes.png")
	png.Encode(f, img)
}
