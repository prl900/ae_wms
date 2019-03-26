package main

import (
	"fmt"
	"log"
	"time"
	"strconv"
	"strings"
	"net/http"

	"github.com/terrascope/geometry"
	"github.com/prl900/ae_wms/tree/geo_array/rastreader"
)

const (
        webMerc = "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext  +no_defs"
        wgs84   = "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs "
)

func wms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("%s", r.URL)

	params := r.URL.Query()
	bboxCoords := strings.Split(params["bbox"][0], ",")
	if len(bboxCoords) != 4 {
		http.Error(w, fmt.Sprintf("Malformed WMS GetMap request"), 400)
		return
	}

	var err error
	pts := make([]float64, 4)
	for i, bboxCoord := range bboxCoords {
		pts[i], err = strconv.ParseFloat(bboxCoord, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Malformed WMS GetMap request: %v", err), 400)
			return
		}
	}

	bbox := geometry.BBox(pts[0], pts[1], pts[2], pts[3])
	if bbox.Area() > 4 {
		http.Error(w, fmt.Sprintf("Too big area: %f", bbox.Area()), 413)
		return
	}

	width, err := strconv.ParseInt(params["width"][0], 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Malformed WMS GetMap request: %v", err), 400)
		return
	}

	height, err := strconv.ParseInt(params["height"][0], 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Malformed WMS GetMap request: %v", err), 400)
		return
	}

	var d time.Time
	out, err := rastreader.GenerateModisTile(int(width), int(height), bbox, d, wgs84)
        if err != nil {
		http.Error(w, fmt.Sprintf("Error reading from object: %v", err), 400)
                return
        }

	fmt.Fprintf(w, "%s", out)
}

func main() {
	http.HandleFunc("/geoarray", wms)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
