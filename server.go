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

	"source.cloud.google.com/wald-1526877012527/cloud_wms/rastreader"

	"github.com/terrascope/geometry"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
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
	if bbox.Area() > 25000000000 {
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
		t, _ = time.Parse(time.RFC3339, params["time"][0])
	} else {
		t, err = time.Parse(time.RFC3339, md[layer].Dates[0])
	}

	paletted, err := rastreader.GenerateTile(md[layer], int(width), int(height), bbox, t)
	if err != nil {
		http.Error(w, fmt.Sprintf("Layer not found"), 400)
		return
	}

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

func prof(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("%s", r.URL)

	params := r.URL.Query()
	arr_name := params["arr"][0]

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating client: %v", err), 400)
		return
	}
	bkt := client.Bucket("wald-1526877012527.appspot.com")

	rdr, err := bkt.Object(fmt.Sprintf("%s.npy", arr_name)).NewReader(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating object reader: %v", err), 400)
		return
	}

	_, err = ioutil.ReadAll(rdr)
	rdr.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading from object: %v", err), 400)
		return
	}

	fmt.Fprintf(w, "Success")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/wms", wms)
	http.HandleFunc("/prof", prof)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
