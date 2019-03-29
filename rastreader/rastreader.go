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
	Modis4tileName = "modis_arr/MCD43A4.A2018001.h%02dv%02d.006_b%s_%d"
	xExtentModis  = 1111950.519666
	yExtentModis  = 1111950.519667
	XSize = 2400
	YSize = 2400
	X4Size = 1200
	Y4Size = 1200
)

// -------------------------- 2400 --------------------------

type Modis4TileID struct {
	Horizontal int
	Vertical   int
	SeqH       int
	SeqV       int
}

func xy24tile(x, y float64) Modis4TileID {
	return Modis4TileID{Horizontal: int(math.Floor(x/xExtentModis)) + 18, Vertical: -1*int(math.Ceil(y/yExtentModis)) + 9,
		SeqH: int((math.Mod(x, xExtentModis) / xExtentModis) * 2), SeqV: int((-1 * (math.Mod(y, yExtentModis) / yExtentModis)) * 2)}
}

func get4Width(a, b Modis4TileID) int {
	return (b.Horizontal-a.Horizontal)*2 - a.SeqH + b.SeqH
}

func get4Height(a, b Modis4TileID) int {
	return (b.Vertical-a.Vertical)*2 - a.SeqV + b.SeqV
}

func ListModis4TileIDs(bbox geometry.BoundingBox, proj4 string, geog bool) []Modis4TileID {
	pts := []geometry.Point{{bbox.Min.X, bbox.Min.Y}, {bbox.Max.X, bbox.Max.Y}}

	if !geog {
		proj4go.Inverse(proj4, pts)
	}

	proj4go.Forwards(sinuProj, pts)

	tlTile := xy24tile(pts[0].X, pts[1].Y)
	brTile := xy24tile(pts[1].X, pts[0].Y)

	seqs := []Modis4TileID{}
	for j := 0; j <= get4Height(tlTile, brTile); j++ {
		for i := 0; i <= get4Width(tlTile, brTile); i++ {
			seqs = append(seqs, Modis4TileID{tlTile.Horizontal + (tlTile.SeqH+i)/2, tlTile.Vertical + (tlTile.SeqV+j)/2, (tlTile.SeqH + i) % 2, (tlTile.SeqV + j) % 2})
		}
	}

	return seqs
}

func GetModis4Info(tile Modis4TileID) *raster.Raster {
	x0 := float64(tile.Horizontal-18)*xExtentModis + float64(tile.SeqH)*(xExtentModis/2)
	x1 := x0 + xExtentModis/2
	y1 := float64(9-tile.Vertical)*yExtentModis - float64(tile.SeqV)*(yExtentModis/2)
	y0 := y1 - yExtentModis/2

	return &raster.Raster{scimage.NewBlankImage(scicolor.GrayU8Model{1, 255, 0}, image.Rect(0, 0, X4Size, Y4Size)),
		proj4go.Coverage{Proj4: sinuProj, BoundingBox: geometry.BBox(x0, y0, x1, y1)}}
}

func ReadModis4Tile(tile Modis4TileID, date time.Time, band string) (*scimage.GrayU8, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v", err)
	}
	bkt := client.Bucket(bucketName)

	// Add bands parameters
	//objName := fmt.Sprintf(ModistileName, tile.Horizontal, tile.Vertical, 2, date.Format("2006.01.02"))
	objName := fmt.Sprintf(ModistileName, tile.Horizontal, tile.Vertical, band, tile.SeqV*10+tile.SeqH)
	r, err := bkt.Object(objName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating object reader: %s object: %s: %v", bucketName, objName, err)
	}

	img, err := png.Decode(r)
        g8 := img.(*image.Gray)

	return &scimage.GrayU8{Pix: g8.Pix, Stride: XSize, Rect: image.Rect(0, 0, XSize, XSize), Min: 1, Max: 255, NoData: 0}, nil
}

func GenerateModis4Tile(width, height int, bbox geometry.BoundingBox, date time.Time, band, proj4 string)  ([]uint8, error) {
	img := scimage.NewGrayU8(image.Rect(0, 0, width, height), 1, 255, 0)
	rMerc := &raster.Raster{Image: img, Coverage: proj4go.Coverage{BoundingBox: bbox, Proj4: proj4}}

	tiles := ListModis4TileIDs(bbox, proj4, false)

	var err error
	for _, tile := range tiles {
		rIn := GetModis4Info(tile)
		rIn.Image, err = ReadModis4Tile(tile, date, band)
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

// -------------------------- 2400 --------------------------

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
