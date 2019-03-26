package main

import (
	"fmt"
	"log"
	"image"
	"image/png"
	"net/http"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const (
        bucketName = "wald-1526877012527.appspot.com"
)

func wms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("%s", r.URL)

	ctx := context.Background()
        client, err := storage.NewClient(ctx)
        if err != nil {
		http.Error(w, fmt.Sprintf("Error reading from object: %v", err), 400)
                return
        }

	bkt := client.Bucket(bucketName)

        objName := "MCD43A4.A2018001.h17v04.006_b2"
        rd, err := bkt.Object(objName).NewReader(ctx)
        if err != nil {
		http.Error(w, fmt.Sprintf("Error reading from object: %v", err), 400)
                return
        }

	img, err := png.Decode(rd)
	g8 := img.(*image.Gray)
	fmt.Fprintf(w, "%s", g8.Pix)
}

func main() {
	http.HandleFunc("/geoarray", wms)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
