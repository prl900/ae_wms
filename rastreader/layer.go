package rastreader

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
)

type Layer struct {
	Name         string        `json:"name"`
	Abstract     string        `json:"abstract"`
	Dates        []string      `json:"dates_iso8601"`
	XSize        int           `json:"x_size"`
	YSize        int           `json:"y_size"`
	Geotransform []float64     `json:"geotransform"`
	MaxVal       float32       `json:"max_value"`
	MinVal       float32       `json:"min_value"`
	NoData       float32       `json:"no_data"`
	Proj4        string        `json:"proj4"`
	Palette      []color.NRGBA `json:"palette"`
}

type Layers map[string]Layer

func ReadLayers(fileName string) (Layers, error) {
	lyrs := Layers{}

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return lyrs, err
	}

	json.Unmarshal(bytes, &lyrs)

	return lyrs, nil
}
