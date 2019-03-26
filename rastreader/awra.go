package rastreader

import (
	"fmt"
	"image"
	"io/ioutil"
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
	wgs84      = "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs "

	e0tileName = "awra_bom/%s_%s.dat.snp"
)


func GetAWRAInfo(layer Layer) *raster.Raster {
	return &raster.Raster{scimage.NewBlankImage(scicolor.GrayF32Model{float32(layer.MinVal), float32(layer.MaxVal), float32(layer.NoData)}, image.Rect(0, 0, layer.XSize, layer.YSize)),
		proj4go.Coverage{Proj4: wgs84, BoundingBox: geometry.BBox(112, -10, 154, -44)}}
}

func ReadAWRATile(layer Layer, date time.Time) (*scimage.GrayF32, error) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v", err)
	}
	bkt := client.Bucket(bucketName)

	objName := fmt.Sprintf(e0tileName, layer.Name, date.Format("20060102"))
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
	header.Len /= 4
	header.Cap /= 4

	pix := *(*[]float32)(unsafe.Pointer(&header))

	return &scimage.GrayF32{Pix: pix, Stride: layer.XSize, Rect: image.Rect(0, 0, layer.XSize, layer.YSize), Min: float32(layer.MinVal), Max: float32(layer.MaxVal), NoData: float32(layer.NoData)}, nil
}

func GenerateAwraTile(layer Layer, width, height int, bbox geometry.BoundingBox, date time.Time)  (*image.Paletted, error) {

	img := scimage.NewGrayF32(image.Rect(0, 0, width, height), layer.MinVal, layer.MaxVal, layer.NoData)
	rMerc := &raster.Raster{Image: img, Coverage: proj4go.Coverage{BoundingBox: bbox, Proj4: webMerc}}

	rIn := GetAWRAInfo(layer)
	rIn.Image, _ = ReadAWRATile(layer, date)

	rMerc.Warp(rIn)

	return img.AsPaletted(scicolor.GradientNRGBAPalette(layer.Palette)), nil

}
