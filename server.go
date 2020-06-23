package main //import source.cloud.google.com/wald-1526877012527/cloud_wms

import (
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/prl900/ae_wms/rastreader"
	"github.com/terrascope/geometry"
)

var md rastreader.Layers

func init() {
	var err error
	md, err = rastreader.ReadLayers("metadata.json")
	if err != nil {
		panic(err)
	}
	fmt.Println(md)
}

func wms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("%s", r.URL)

	params := r.URL.Query()
	if params["request"][0] == "GetCapabilities" {
		ExecuteWriteTemplateFile(w, md, "templates/WMS_GetCapabilities.tpl")
		return
	}

	if params["service"][0] != "WMS" || params["request"][0] != "GetMap" || params["srs"][0] != "EPSG:3857" {
		http.Error(w, fmt.Sprintf("Malformed WMS GetMap request"), 400)
		return
	}

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

	if bbox.Area() > 4e11 {
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

	layer := strings.Split(params["layers"][0], ",")[0]

	var t time.Time
	if _, ok := params["time"]; ok {
		contains := false
		for _, d := range md[layer].Dates {
			if d == params["time"][0] {
				contains = true
				break
			}
		}
		if !contains {
			http.Error(w, fmt.Sprintf("Malformed WMS GetMap request: %s in not defined in the server. Available dates are: %v", params["time"][0], md[layer].Dates), 400)
			return
		}
		t, _ = time.Parse(time.RFC3339, params["time"][0])
	} else {
		t, err = time.Parse(time.RFC3339, md[layer].Dates[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not parse date in metadata file %v", err), 500)
			return
		}
	}

	paletted, err := rastreader.GenerateTile(md[layer], int(width), int(height), bbox, t)
	if err != nil {
		http.Error(w, fmt.Sprintf("Layer not found %v", err), 400)
		return
	}

	// Enable browser and intermediate caching
	w.Header().Set("Cache-Control", "public, max-age:31536000")

	err = png.Encode(w, paletted)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error PNG encoding tile: %v", err), 400)
		return
	}
}

func ExecuteWriteTemplateFile(w io.Writer, data interface{}, filePath string) error {
	tplStr, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error trying to read %s file: %v", filePath, err)
	}
	tpl, err := template.New("template").Parse(string(tplStr))
	if err != nil {
		return fmt.Errorf("Error trying to parse template document: %v", err)
	}
	err = tpl.Execute(w, data)
	if err != nil {
		return fmt.Errorf("Error executing template: %v\n", err)
	}

	return nil
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/wms", wms)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
