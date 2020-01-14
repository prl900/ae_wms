package rastreader

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/terrascope/geometry"
	geo "github.com/terrascope/geometry"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"

	"cloud.google.com/go/storage"
	"github.com/golang/snappy"
	"golang.org/x/net/context"
)

const (
	tileName = "dea/fc_metrics_maxPV_%+04d_%+04d_l%d_2001.snp"
	gda94    = "+proj=aea +lat_1=-18 +lat_2=-36 +lat_0=0 +lon_0=132 +x_0=0 +y_0=0 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs "
)

func WarpTile(x, y, level int, out *raster.Raster, wg *sync.WaitGroup) error {
	defer wg.Done()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	bkt := client.Bucket(bucketName)

	tileStep := (1 << level)
	tileCov := proj4go.Coverage{BoundingBox: geo.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: gda94}

	objName := fmt.Sprintf(tileName, x, y, level)
	r, err := bkt.Object(objName).NewReader(ctx)
	if err != nil {
		return nil
	}

	cdata, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return fmt.Errorf("Error reading from object: %s object: %s: %v", bucketName, objName, err)
	}

	data, err := snappy.Decode(nil, cdata)
	if err != nil {
		return fmt.Errorf("Error decompressing data: %s object: %s: %v", bucketName, objName, err)
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
	case res > 800:
		level = 5
	case res > 400:
		level = 4
	case res > 200:
		level = 3
	case res > 100:
		level = 2
	case res > 50:
		//level = 1
		level = 2
	default:
		//level = 0
		level = 2
	}

	tileStep := (1 << level)

	covGDA94, err := cov.Transform(gda94)
	if err != nil {
		return nil, err
	}

	minX := int(covGDA94.BoundingBox.Min.X / 1e4)
	minY := int(covGDA94.BoundingBox.Min.Y / 1e4)
	maxX := int(covGDA94.BoundingBox.Max.X / 1e4)
	maxY := int(covGDA94.BoundingBox.Max.Y / 1e4)

	x0 := (minX+190)/tileStep*tileStep - 190
	x1 := (((maxX+190)/tileStep)+1)*tileStep - 190
	y0 := ((minY+100)/tileStep-1)*tileStep - 100
	y1 := ((maxY+100)/tileStep)*tileStep - 100

	var wg sync.WaitGroup

	for x := x0; x <= x1; x += tileStep {
		for y := y1; y >= y0; y -= tileStep {
			wg.Add(1)
			go WarpTile(x, y, level, rMerc, &wg)
			/*
				tileCov := proj4go.Coverage{BoundingBox: geo.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: gda94}

				objName := fmt.Sprintf(tileName, x, y, level)
				r, err := bkt.Object(objName).NewReader(ctx)
				if err != nil {
					continue
				}

				cdata, err := ioutil.ReadAll(r)
				r.Close()
				if err != nil {
					return nil, fmt.Errorf("Error reading from object: %s object: %s: %v", bucketName, objName, err)
				}

				data, err := snappy.Decode(nil, cdata)
				if err != nil {
					return nil, fmt.Errorf("Error decompressing data: %s object: %s: %v", bucketName, objName, err)
				}

				im := &scimage.GrayU8{Pix: data, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: 0, Max: 100, NoData: 255}
				rIn := &raster.Raster{im, tileCov}
				rMerc.Warp(rIn)
			*/
		}
	}

	wg.Wait()

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
