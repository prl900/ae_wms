package rastreader

import (
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"sync"
	"time"
	"os"


	"github.com/terrascope/geometry"
	geo "github.com/terrascope/geometry"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"

	"github.com/golang/snappy"
)

const (
	tileName = "/home/p_rozas_larraondo/data/fc_metrics_WCF_%+04d_%+04d_l%d_%d.snp"
	gda94    = "+proj=aea +lat_1=-18 +lat_2=-36 +lat_0=0 +lon_0=132 +x_0=0 +y_0=0 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs "
)

func WarpTile(x, y, level, year int, out *raster.Raster, wg *sync.WaitGroup) error {
	defer wg.Done()

	tileStep := (1 << level)
	tileCov := proj4go.Coverage{BoundingBox: geo.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: gda94}
	fmt.Println(tileCov)

	fName := fmt.Sprintf(tileName, x, y, level, year)
	fmt.Println(fName)
	r, err := os.Open(fName)
	if err != nil {
		fmt.Println("Hmmm:", err)
		return nil
	}

	cdata, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return fmt.Errorf("Error reading from object: %s object: %s: %v", fName, err)
	}

	data, err := snappy.Decode(nil, cdata)
	if err != nil {
		return fmt.Errorf("Error decompressing data: %s object: %s: %v", fName, err)
	}

	im := &scimage.GrayU8{Pix: data, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: 0, Max: 100, NoData: 255}
	rIn := &raster.Raster{im, tileCov}
	out.Warp(rIn)

	return nil

}

func GenerateDEATile(layer Layer, width, height int, bbox geometry.BoundingBox, date time.Time) (*image.Paletted, error) {
	img := scimage.NewGrayU8(image.Rect(0, 0, width, height), 0, 100, 255)
	cov := proj4go.Coverage{BoundingBox: bbox, Proj4: webMerc}
	rMerc := &raster.Raster{Image: img, Coverage: cov}

	var level int
	switch res := rMerc.Resolution()[0]; {
	case res > 400:
		level = 5
	case res > 200:
		level = 4
	case res > 100:
		level = 3
	case res > 50:
		level = 2
	case res > 25:
		level = 1
	default:
		level = 1
	}

	tileStep := (1 << level)

	covGDA94, err := cov.Transform(gda94)
	if err != nil {
		return nil, err
	}

	minX := int(math.Floor(covGDA94.BoundingBox.Min.X / 1e4))
	minY := int(math.Floor(covGDA94.BoundingBox.Min.Y / 1e4))
	maxX := int(math.Ceil(covGDA94.BoundingBox.Max.X / 1e4))
	maxY := int(math.Ceil(covGDA94.BoundingBox.Max.Y / 1e4))

	x0 := (minX+190)/tileStep*tileStep - 190
	x1 := (maxX+190)/tileStep*tileStep - 190
	y0 := (minY+100)/tileStep*tileStep - 100
	y1 := (maxY+100)/tileStep*tileStep - 100

	var wg sync.WaitGroup

	for x := x0; x <= x1; x += tileStep {
		for y := y1; y >= y0; y -= tileStep {
			wg.Add(1)
			go WarpTile(x, y, level, date.Year(), rMerc, &wg)
		}
	}

	wg.Wait()

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil
}
