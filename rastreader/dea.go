package rastreader

import (
	"fmt"
	"image"
	//"log"
	//"math"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/snappy"
	"github.com/terrascope/geometry"
	geo "github.com/terrascope/geometry"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"
	//"cloud.google.com/go/storage"
	//"golang.org/x/net/context"
)

// Lons: 41
// [-19, -18, -17, -16, -15, -14, -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21]

// Lats: 39
// [-49, -48, -47, -46, -45, -44, -43, -42, -41, -40, -39, -38, -37, -36, -35, -34, -33, -32, -31, -30, -29, -28, -27, -26, -25, -24, -23, -22, -21, -20, -19, -18, -17, -16, -15, -14, -13, -12, -11]

const (
	tileName = "/home/prl900/Downloads/dea_blobs/fc_metrics_maxPV_%+04d_%+04d_l%d_2001.snp"
	gda94    = "+proj=aea +lat_1=-18 +lat_2=-36 +lat_0=0 +lon_0=132 +x_0=0 +y_0=0 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs "
)

func GenerateDEATile(layer Layer, width, height int, bbox geometry.BoundingBox, date time.Time) (*image.Paletted, error) {
	/*
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		bkt := client.Bucket(bucketName)
	*/

	img := scimage.NewGrayU8(image.Rect(0, 0, width, height), 0, 100, 255)
	cov := proj4go.Coverage{BoundingBox: bbox, Proj4: webMerc}
	rMerc := &raster.Raster{Image: img, Coverage: cov}

	var level int
	/*
		switch res := rMerc.Resolution()[0]; {
		case res > 800:
			level = 5
		case res > 400:
			level = 4
		case res > 200:
			level = 3
		case res > 100:
			level = 2
		case res > 50:
			level = 1
		default:
			level = 0
		}
	*/
	level = 3
	tileStep := (1 << level) // * 10
	fmt.Println(tileStep)

	covGDA94, err := cov.Transform(gda94)
	if err != nil {
		return nil, err
	}

	fmt.Println(covGDA94.BoundingBox)

	minX := int(covGDA94.BoundingBox.Min.X / 1e4)
	minY := int(covGDA94.BoundingBox.Min.Y / 1e4)
	maxX := int(covGDA94.BoundingBox.Max.X / 1e4)
	maxY := int(covGDA94.BoundingBox.Max.Y / 1e4)

	x0 := (minX+190)/tileStep*tileStep - 190
	x1 := (((maxX+190)/tileStep)+1)*tileStep - 190
	y0 := ((minY+100)/tileStep)*tileStep - 100
	y1 := (((maxY+100)/tileStep)-1)*tileStep - 100

	fmt.Println(minX, maxX, minY, maxY)
	fmt.Println(x0, x1, y0, y1)

	for x := x0; x <= x1; x += tileStep {
		for y := y1; y >= y0; y -= tileStep {
			tileCov := proj4go.Coverage{BoundingBox: geo.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: gda94}
			fmt.Printf(tileName, x, y, level)
			fmt.Println()
			fmt.Println(tileCov.BoundingBox)
			fName := fmt.Sprintf(tileName, x, y, level)
			if _, err := os.Stat(fName); err != nil {
				fmt.Println("No")
				continue
			}

			data, _ := ioutil.ReadFile(fName)
			cdata, _ := snappy.Decode(nil, data)

			im := &scimage.GrayU8{Pix: cdata, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: 0, Max: 100, NoData: 255}
			rIn := &raster.Raster{im, tileCov}
			rMerc.Warp(rIn)
		}
	}

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil
}

/*
	im, err := cog.DecodeLevelSubImage(rc, level, image.Rect(i0, j0, i1, j1))
	rIn := &raster.Raster{im, proj4go.Coverage{BoundingBox: geo.BBox(x0, y0, x1, y1), Proj4: gda94}}

	rMerc.Warp(rIn)
	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil
}
*/

/*

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			objName := fmt.Sprintf(tileName, x, y, date.Format("2006"))
			rc, err := bkt.Object(objName).NewReader(ctx)
			//rc, err := os.Open(fName)
			if err != nil {
				continue
			}

			i0 := int(math.Floor(math.Max(0, (covGDA94.BoundingBox.Min.X-float64(x)*100000)/1e5) * (4000 / math.Pow(2, float64(level)))))
			j1 := int(4000/math.Pow(2, float64(level))) - int(math.Floor(math.Max(0, (covGDA94.BoundingBox.Min.Y-float64(y)*100000)/1e5)*(4000/math.Pow(2, float64(level)))))
			i1 := int(math.Ceil(math.Min(1, (covGDA94.BoundingBox.Max.X-float64(x)*100000)/1e5) * (4000 / math.Pow(2, float64(level)))))
			j0 := int(4000/math.Pow(2, float64(level))) - int(math.Ceil(math.Min(1, (covGDA94.BoundingBox.Max.Y-float64(y)*100000)/1e5)*(4000/math.Pow(2, float64(level)))))

			x0 := float64(x)*100000 + float64(i0)*25*math.Pow(2, float64(level))
			x1 := float64(x)*100000 + float64(i1)*25*math.Pow(2, float64(level))
			y0 := float64(y+1)*100000 - float64(j0)*25*math.Pow(2, float64(level))
			y1 := float64(y+1)*100000 - float64(j1)*25*math.Pow(2, float64(level))

			im, err := cog.DecodeLevelSubImage(rc, level, image.Rect(i0, j0, i1, j1))
			if err != nil {
				return nil, err
			}

			rIn := &raster.Raster{im, proj4go.Coverage{BoundingBox: geo.BBox(x0, y0, x1, y1), Proj4: gda94}}

			rMerc.Warp(rIn)
		}
	}

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil

}
*/
