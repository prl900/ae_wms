package rastreader

import (
	"context"
	"fmt"
	"image"
	"image/png"
	//"io/ioutil"
	"math"
	//"sync"
	//"strconv"
	"time"

	"github.com/terrascope/geometry"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"

	"cloud.google.com/go/storage"

	"golang.org/x/sync/errgroup"
)

const (
	bucketName = "wald-wms"
	tileName   = "/home/prl900/Downloads/irr_data/irr_DEA_%+04d_%+04d_201702_l%d.png"
)

func ComputeKc(im *image.NRGBA, layer Layer) (*scimage.GrayU8, error) {
	out := scimage.NewGrayU8(im.Rect, uint8(layer.MinVal), uint8(layer.MaxVal), uint8(layer.NoData))

	for i := 0; i < im.Bounds().Dx(); i++ {
		for j := 0; j < im.Bounds().Dy(); j++ {
			c := im.NRGBAAt(i, j)
			red, blue, nir, swir1 := float64(c.R)*2e-3, float64(c.G)*2e-3, float64(c.B)*2e-3, float64(c.A)*2e-3

			evi := 2.5 * (nir - red) / (nir + 6*red - 7.5*blue + 1)
			gvmi := ((nir + 0.1) - (swir1 + 0.02)) / ((nir + 0.1) + (swir1 + 0.02))
			rmi := math.Max(0, gvmi-(0.775*evi-0.0757))
			evir := math.Max(0, math.Min(1, evi))

			kc := 0.680 * (1 - math.Exp(-14.12*math.Pow(evir, 2.482)-7.991*math.Pow(rmi, 0.890)))

			out.SetGrayU8(i, j, scicolor.GrayU8{Y: uint8(kc * 255)})
		}
	}

	return out, nil
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
		level = 0
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

	grp, ctx := errgroup.WithContext(ctx)
	for x := x0; x <= x1; x += tileStep {
		for y := y1; y >= y0; y -= tileStep {
			x, y := x, y
			grp.Go(func() error {
				//objName := fmt.Sprintf(tileName, x, y, level, date.Year())
				objName := fmt.Sprintf(tileName, x, y, level)
				fmt.Println(objName)

				rc, err := bkt.Object(objName).NewReader(ctx)
				if err != nil {
					return nil
				}
				defer rc.Close()

				/*
					cdata, err := ioutil.ReadAll(rc)
					if err != nil {
						return err
					}



					f, err := os.Open(objName)
					if err != nil {
						return nil
					}
					defer f.Close()
				*/

				img, err := png.Decode(rc)
				if err != nil {
					return nil
				}

				im, _ := ComputeKc(img.(*image.NRGBA), layer)

				tileStep := (1 << level)
				tileCov := proj4go.Coverage{BoundingBox: geometry.BBox(float64(x)*1e4, float64(y-tileStep)*1e4, float64(x+tileStep)*1e4, float64(y)*1e4), Proj4: layer.Proj4}

				rIn := &raster.Raster{im, tileCov}

				return rMerc.Warp(rIn)
			})
		}
	}

	if err := grp.Wait(); err != nil {
		return nil, err
	}

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil
}
