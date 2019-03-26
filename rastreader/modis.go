package rastreader

import (
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"reflect"
	"time"
	"unsafe"

	"github.com/terrascope/geometry"
	"github.com/terrascope/proj4go"
	"github.com/terrascope/raster"
	"github.com/terrascope/scimage"
	"github.com/terrascope/scimage/scicolor"

	"cloud.google.com/go/storage"
	"github.com/golang/snappy"
	"golang.org/x/net/context"
)

const (
	sinuProj   = "+proj=sinu +lon_0=0 +x_0=0 +y_0=0 +a=6371007.181 +b=6371007.181 +units=m +no_defs "

	ModistileName = "modis/%s_h%02dv%02ds%02d.%s.snp"
	xExtentModis  = 1111950.519666
	yExtentModis  = 1111950.519667
)

type ModisTileID struct {
	Horizontal int
	Vertical   int
	SeqH       int
	SeqV       int
}

func xy2tile(x, y float64) ModisTileID {
	return ModisTileID{Horizontal: int(math.Floor(x/xExtentModis)) + 18, Vertical: -1*int(math.Ceil(y/yExtentModis)) + 9,
		SeqH: int((math.Mod(x, xExtentModis) / xExtentModis) * 6), SeqV: int((-1 * (math.Mod(y, yExtentModis) / yExtentModis)) * 6)}
}

func getWidth(a, b ModisTileID) int {
	return (b.Horizontal-a.Horizontal)*6 - a.SeqH + b.SeqH
}

func getHeight(a, b ModisTileID) int {
	return (b.Vertical-a.Vertical)*6 - a.SeqV + b.SeqV
}

func ListModisTileIDs(bbox geometry.BoundingBox, geog bool) []ModisTileID {
	pts := []geometry.Point{{bbox.Min.X, bbox.Min.Y}, {bbox.Max.X, bbox.Max.Y}}

	if !geog {
		proj4go.Inverse(webMerc, pts)
	}

	proj4go.Forwards(sinuProj, pts)

	tlTile := xy2tile(pts[0].X, pts[1].Y)
	brTile := xy2tile(pts[1].X, pts[0].Y)

	seqs := []ModisTileID{}
	for j := 0; j <= getHeight(tlTile, brTile); j++ {
		for i := 0; i <= getWidth(tlTile, brTile); i++ {
			seqs = append(seqs, ModisTileID{tlTile.Horizontal + (tlTile.SeqH+i)/6, tlTile.Vertical + (tlTile.SeqV+j)/6, (tlTile.SeqH + i) % 6, (tlTile.SeqV + j) % 6})
		}
	}

	return seqs
}

func GetModisInfo(layer Layer, tile ModisTileID) *raster.Raster {
	x0 := float64(tile.Horizontal-18)*xExtentModis + float64(tile.SeqH)*(xExtentModis/6)
	x1 := x0 + xExtentModis/6
	y1 := float64(9-tile.Vertical)*yExtentModis - float64(tile.SeqV)*(yExtentModis/6)
	y0 := y1 - yExtentModis/6

	return &raster.Raster{scimage.NewBlankImage(scicolor.GrayS16Model{int16(layer.MinVal), int16(layer.MaxVal), int16(layer.NoData)}, image.Rect(0, 0, layer.XSize, layer.YSize)),
		proj4go.Coverage{Proj4: sinuProj, BoundingBox: geometry.BBox(x0, y0, x1, y1)}}
}

func ReadModisTile(layer Layer, tile ModisTileID, date time.Time) (*scimage.GrayS16, error) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v", err)
	}
	bkt := client.Bucket(bucketName)

	objName := fmt.Sprintf(ModistileName, layer.Name, tile.Horizontal, tile.Vertical, tile.SeqV*10+tile.SeqH, date.Format("2006.01.02"))
	r, err := bkt.Object(objName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating object reader: %s object: %s: %v", bucketName, objName, err)
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
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&data))
	header.Len /= 2
	header.Cap /= 2
	pix := *(*[]int16)(unsafe.Pointer(&header))

	return &scimage.GrayS16{Pix: pix, Stride: layer.XSize, Rect: image.Rect(0, 0, layer.XSize, layer.YSize), Min: int16(layer.MinVal), Max: int16(layer.MaxVal), NoData: int16(layer.NoData)}, nil
}

func GenerateModisTile(layer Layer, width, height int, bbox geometry.BoundingBox, date time.Time)  (*image.Paletted, error) {
	img := scimage.NewGrayS16(image.Rect(0, 0, width, height), int16(layer.MinVal), int16(layer.MaxVal), int16(layer.NoData))
	rMerc := &raster.Raster{Image: img, Coverage: proj4go.Coverage{BoundingBox: bbox, Proj4: webMerc}}

	tiles := ListModisTileIDs(bbox, false)

	var err error
	for _, tile := range tiles {
		rIn := GetModisInfo(layer, tile)
		rIn.Image, err = ReadModisTile(layer, tile, date)
		if err != nil {
			continue
		}

		rMerc.Warp(rIn)
	}

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil
}
