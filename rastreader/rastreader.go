package rastreader

import (
	"fmt"
	"image"
	"time"

	"github.com/terrascope/geometry"
)

const (
	bucketName = "wald-1526877012527.appspot.com"
	webMerc    = "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext  +no_defs"
)

func GenerateTile(layer Layer, width, height int, bbox geometry.BoundingBox, date time.Time) (*image.Paletted, error) {

	switch layer.Name {
	case "ndvi" ,"cmrset":
		return GenerateModisTile(layer, width, height, bbox, date)
	case "e0":
		return GenerateAwraTile(layer, width, height, bbox, date)

	default:
		return nil, fmt.Errorf("Layer not found")
	}

}
