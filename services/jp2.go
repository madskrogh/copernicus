package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/appengine/urlfetch"
)

//getColour calls to the jpeg200 analysis api (https://github.com/jf87/jp2-service)
//passing the given given image path and recieving/
//returning the colour information for the image
func GetColour(path string, rlevel int, ctx *context.Context) (*[][]int, error) {
	trimmedPath := strings.TrimPrefix(path, "console.cloud.google.com/storage/browser/gcp-public-data-sentinel-2/")
	postMap := map[string]interface{}{"path": trimmedPath, "rlevel": rlevel}
	postJSON, err := json.Marshal(postMap)
	if err != nil {
		return nil, err
	}
	client := urlfetch.Client(*ctx)
	req, err := http.NewRequest("POST", "http://35.227.24.82/api/jp2", bytes.NewBuffer(postJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	type Image struct {
		Colours        [][]int `json:"img_data"`
		Shape          []int   `json:"shape"`
		TimeDownload   float64 `json:"time_download"`
		TimeProcessing float64 `json:"time_processing"`
	}
	var img Image
	err = json.Unmarshal(body, &img)
	if err != nil {
		return nil, err
	}
	return &img.Colours, nil
}
