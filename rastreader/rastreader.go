package rastreader

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"time"

	"github.com/terrascope/geometry"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const (
	bucketName = "wald-1526877012527.appspot.com"
	sinuProj   = "+proj=sinu +lon_0=0 +x_0=0 +y_0=0 +a=6371007.181 +b=6371007.181 +units=m +no_defs "

	ModistileName = "modis_arr/MCD43A4.A2018001.h%02dv%02d.006_b%s"
	xExtentModis  = 1111950.519666
	yExtentModis  = 1111950.519667
	XSize = 2400
	YSize = 2400
)

type ModisTileID struct {
	Horizontal int
	Vertical   int
}

func xy2tile(x, y float64) ModisTileID {
	return ModisTileID{Horizontal: int(math.Floor(x/xExtentModis)) + 18, Vertical: -1*int(math.Ceil(y/yExtentModis)) + 9}
}

func getWidth(a, b ModisTileID) int {
	return (b.Horizontal-a.Horizontal)
}

func getHeight(a, b ModisTileID) int {
	return (b.Vertical-a.Vertical)
}

func ListModisTileIDs(bbox geometry.BoundingBox, proj4 string, geog bool) []ModisTileID {
	pts := []geometry.Point{{bbox.Min.X, bbox.Min.Y}, {bbox.Max.X, bbox.Max.Y}}

	if !geog {
		proj4go.Inverse(proj4, pts)
	}

	proj4go.Forwards(sinuProj, pts)

	tlTile := xy2tile(pts[0].X, pts[1].Y)
	brTile := xy2tile(pts[1].X, pts[0].Y)

	seqs := []ModisTileID{}
	for j := 0; j <= getHeight(tlTile, brTile); j++ {
		for i := 0; i <= getWidth(tlTile, brTile); i++ {
			seqs = append(seqs, ModisTileID{tlTile.Horizontal + i, tlTile.Vertical + j})
		}
	}

	return seqs
}

func GetModisInfo(tile ModisTileID) *raster.Raster {
	x0 := float64(tile.Horizontal-18)*xExtentModis
	x1 := x0 + xExtentModis
	y1 := float64(9-tile.Vertical)*yExtentModis
	y0 := y1 - yExtentModis

	return &raster.Raster{scimage.NewBlankImage(scicolor.GrayU8Model{1, 255, 0}, image.Rect(0, 0, XSize, YSize)),
		proj4go.Coverage{Proj4: sinuProj, BoundingBox: geometry.BBox(x0, y0, x1, y1)}}
}

func ReadModisTile(tile ModisTileID, date time.Time, band string) (*scimage.GrayU8, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v", err)
	}
	bkt := client.Bucket(bucketName)

	// Add bands parameters
	//objName := fmt.Sprintf(ModistileName, tile.Horizontal, tile.Vertical, 2, date.Format("2006.01.02"))
	objName := fmt.Sprintf(ModistileName, tile.Horizontal, tile.Vertical, band)
	r, err := bkt.Object(objName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating object reader: %s object: %s: %v", bucketName, objName, err)
	}

	img, err := png.Decode(r)
        g8 := img.(*image.Gray)

	return &scimage.GrayU8{Pix: g8.Pix, Stride: XSize, Rect: image.Rect(0, 0, XSize, XSize), Min: 1, Max: 255, NoData: 0}, nil
}

func GenerateModisTile(width, height int, bbox geometry.BoundingBox, date time.Time, band, proj4 string)  ([]uint8, error) {
	img := scimage.NewGrayU8(image.Rect(0, 0, width, height), 1, 255, 0)
	rMerc := &raster.Raster{Image: img, Coverage: proj4go.Coverage{BoundingBox: bbox, Proj4: proj4}}

	tiles := ListModisTileIDs(bbox, proj4, false)

	var err error
	for _, tile := range tiles {
		rIn := GetModisInfo(tile)
		rIn.Image, err = ReadModisTile(tile, date, band)
		if err != nil {
			fmt.Println("Error!", err)
			continue
		}

		err := rMerc.Warp(rIn)
		if err != nil {
			return nil, err
		}

	}

	return img.Pix, nil
}
