package main

import (
	"fmt"
	"os"
	"github.com/terrascope/geometry"
	"image"
	"image/png"
	"github.com/prl900/ae_wms/tree/geo_array/rastreader"
	"time"
)

const (
        webMerc = "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext  +no_defs"
        wgs84   = "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs "
)


func main() {
	bbox := geometry.BBox(0, 0, 10, 10)
	if bbox.Area() > 25000000000 {
		fmt.Println("Too big")
		return
	}

	var d time.Time

	out, err := rastreader.GenerateModisTile(256, 256, bbox, d, wgs84)
	fmt.Println(err)
	fmt.Println(out[0:100])
	fmt.Println(len(out))

	img := &image.Gray{out, 256, image.Rect(0, 0, 256, 256)}

	f, _ := os.Create("./yes.png")
	png.Encode(f, img)
}
