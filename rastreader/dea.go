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
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"

	"github.com/golang/snappy"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const (
	//tileName = "/home/p_rozas_larraondo/data/fc_metrics_WCF_%+04d_%+04d_l%d_%d.snp"
	//tileName = "/home/p_rozas_larraondo/data/irr_water_%+04d_%+04d_l%d.snp"
	bucketName = "wald-wms"
	//tileName = "irr_water_%+04d_%+04d_l%d.snp"
	tileName = "fc_metrics_WCF_%+04d_%+04d_l%d_%d.snp"
)

func DrillTile(x, y, level int, poly *geometry.Polygon, layer Layer, wg *sync.WaitGroup) (Stat, error) {
	defer wg.Done()

	stat := Stat{}

	tileStep := (1 << level)
	tileCov := proj4go.Coverage{BoundingBox: geometry.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: layer.Proj4}

	fName := fmt.Sprintf(tileName, x, y, level)
	r, err := os.Open(fName)
	if err != nil {
		fmt.Println(fName, ": not found")
		return stat, nil
	}

	cdata, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return stat, fmt.Errorf("Error reading from object: %s object: %s: %v", fName, err)
	}

	data, err := snappy.Decode(nil, cdata)
	if err != nil {
		return stat, fmt.Errorf("Error decompressing data: %s object: %s: %v", fName, err)
	}

	//im := &scimage.GrayU8{Pix: data, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: 0, Max: 100, NoData: 255}
	im := &scimage.GrayU8{Pix: data, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: uint8(layer.MinVal), Max: uint8(layer.MaxVal), NoData: uint8(layer.NoData)}
	rIn := &raster.Raster{im, tileCov}

	rIn.CropPolygon(*poly)

	for _, val := range im.Pix {
		if val != im.NoData {
			stat.Sum += float64(val)
			stat.Count += 1
		}
	}

	return stat, nil

}

func DrillDEA(layer Layer, poly geometry.Polygon) (string, error) {

	cov := proj4go.Coverage{BoundingBox: poly.BoundingBox(), Proj4: geographic}
	covNat, err := cov.Transform(layer.Proj4)
	if err != nil {
		return "", err
	}

	minX := int(math.Floor(covNat.BoundingBox.Min.X / 1e4))
	minY := int(math.Floor(covNat.BoundingBox.Min.Y / 1e4))
	maxX := int(math.Ceil(covNat.BoundingBox.Max.X / 1e4))
	maxY := int(math.Ceil(covNat.BoundingBox.Max.Y / 1e4))

	level := 0
	tileStep := (1 << level)

	x0 := (minX+190)/tileStep*tileStep - 190
	x1 := (maxX+190)/tileStep*tileStep - 190
	y0 := (minY+100)/tileStep*tileStep - 100
	y1 := (maxY+100)/tileStep*tileStep - 100

	g := proj4go.ProjGeometry{&poly, geographic}
	g, err = g.Transform(layer.Proj4)
	if err != nil {
		fmt.Println("Error reprojecting tile")
		return "", err
	}

	p := g.Geometry.(*geometry.Polygon)

	var wg sync.WaitGroup

	for x := x0; x <= x1; x += tileStep {
		for y := y1; y >= y0; y -= tileStep {
			wg.Add(1)
			go DrillTile(x, y, level, p, layer, &wg)
		}
	}

	wg.Wait()

	return fmt.Sprintf("%d, %d, %d, %d", x0, x1, y0, y1), nil
}

func WarpTile(x, y, level, year int, out *raster.Raster, layer Layer, ctx context.Context, bkt *storage.BucketHandle, wg *sync.WaitGroup) error {
	defer wg.Done()

	tileStep := (1 << level)
	tileCov := proj4go.Coverage{BoundingBox: geometry.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: layer.Proj4}

	objName := fmt.Sprintf(tileName, x, y, level, year)
	fmt.Println(objName)
	rc, err := bkt.Object(objName).NewReader(ctx)
	if err != nil {
		fmt.Println(objName, ": not found", err)
		return nil
	}
	defer rc.Close()

	cdata, err := ioutil.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("Error reading from object: %s object: %s: %v", objName, err)
	}

	data, err := snappy.Decode(nil, cdata)
	if err != nil {
		return fmt.Errorf("Error decompressing data: %s object: %s: %v", objName, err)
	}

	//im := &scimage.GrayU8{Pix: data, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: 0, Max: 100, NoData: 255}
	im := &scimage.GrayU8{Pix: data, Stride: 400, Rect: image.Rect(0, 0, 400, 400), Min: uint8(layer.MinVal), Max: uint8(layer.MaxVal), NoData: uint8(layer.NoData)}
	rIn := &raster.Raster{im, tileCov}
	out.Warp(rIn)

	return nil

}

func GenerateDEATile(layer Layer, width, height int, bbox geometry.BoundingBox, date time.Time) (*image.Paletted, error) {

	img := scimage.NewGrayU8(image.Rect(0, 0, width, height), uint8(layer.MinVal), uint8(layer.MaxVal), uint8(layer.NoData))
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

	covGDA94, err := cov.Transform(layer.Proj4)
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

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	bkt := client.Bucket(bucketName)


	var wg sync.WaitGroup

	for x := x0; x <= x1; x += tileStep {
		for y := y1; y >= y0; y -= tileStep {
			wg.Add(1)
			go WarpTile(x, y, level, date.Year(), rMerc, layer, ctx, bkt, &wg)
		}
	}

	wg.Wait()

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil
}
