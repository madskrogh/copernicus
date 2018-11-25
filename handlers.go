package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/madskrogh/copernicus/services"
	"github.com/madskrogh/copernicus/utils"
	"google.golang.org/appengine"
)

func init() {
	http.HandleFunc("/rgbrank", rgbColourDistanceHandler)
	http.HandleFunc("/brank", blueColourDistanceHandler)
	http.HandleFunc("/coordinates", coordinatesHandler)
	http.HandleFunc("/morecoordinates", moreCoordinatesHandler)
	http.HandleFunc("/address", addressHandler)
}

//2.1
//For a given set of coordinates, returns the paths of
//the blue, red and green of the three most recent images.
func coordinatesHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	type RequestVals struct {
		Lon float64
		Lat float64
	}
	lon, err := strconv.ParseFloat(req.FormValue("Lon"), 64)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	lat, err := strconv.ParseFloat(req.FormValue("Lat"), 64)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	reqvals := RequestVals{lon, lat}
	if reqvals.Lon == 0 || reqvals.Lat == 0 {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(req)
	images, err := services.GetImages(reqvals.Lon, reqvals.Lat, &ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	imagePaths := utils.GetImagePaths(images)
	e := json.NewEncoder(w)
	e.Encode(imagePaths)
}

//3.1
//For a given adress, returns the paths of the
//blue, red and green of the three most recent images.
func addressHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	type RequestVals struct {
		Address string
	}
	reqvals := RequestVals{req.FormValue("Address")}
	ctx := appengine.NewContext(req)
	results, err := services.GetAddress(reqvals.Address, &ctx)
	if err != nil {
		http.Error(w, "Address was not found", http.StatusNotFound)
		return
	}
	images, err := services.GetImages(results.Geometry.Location.Lng, results.Geometry.Location.Lat, &ctx)
	if err != nil {
		http.Error(w, "Images not found", http.StatusNotFound)
		return
	}
	imagePaths := utils.GetImagePaths(images)
	e := json.NewEncoder(w)
	e.Encode(imagePaths)
}

//3.2a
//For a given location, returns the paths to the blue
//bands of the three most recent images, ranked in
//increasing order by the distance to a colour value of 255.
func blueColourDistanceHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	type RequestVals struct {
		Address string
	}
	reqvals := RequestVals{req.FormValue("Address")}
	ctx := appengine.NewContext(req)
	results, err := services.GetAddress(reqvals.Address, &ctx)
	if err != nil {
		http.Error(w, "Address was not found:", http.StatusNotFound)
		return
	}
	images, err := services.GetImages(results.Geometry.Location.Lng, results.Geometry.Location.Lat, &ctx)
	if err != nil {
		http.Error(w, "Images not found", http.StatusNotFound)
		return
	}
	type Result struct {
		Message int
		Error   error
	}
	imagePaths := utils.GetImagePaths(images)
	rlevel := -1
	avgs := make([]int, 3)
	outs := make([]chan Result, 0)
	avgMap := make(map[int]string)
	count := 0
	ctx1 := appengine.NewContext(req)
	//for each image blue band path, create a new
	//channel, append/store it and pass it to the
	//go func which makes concurrent calls to the
	//jpeg2000 analysis api, and then the avg function
	//returning the result to the appropriate channel
	for i := 0; i < len(imagePaths); i += 3 {
		out := make(chan Result)
		outs = append(outs, out)
		go func(path string, rlevel int, out chan Result) {
			colour, err := services.GetColour(path, rlevel, &ctx1)
			if err != nil {
				colour, err = services.GetColour(path, rlevel, &ctx1)
				if err != nil {
					out <- Result{0, err}
				}
			}
			out <- Result{int(math.Abs(float64((utils.Avg(colour) - 255)))), nil}
		}(imagePaths[i], rlevel, out)
		count++
	}
	//the values returned by the go routines are
	//channeled to the avgs array
	for i := range outs {
		result := <-outs[i]
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusBadGateway)
			return
		}
		avgs[i] = result.Message
	}
	avgMap[avgs[0]] = imagePaths[0]
	avgMap[avgs[1]] = imagePaths[3]
	avgMap[avgs[2]] = imagePaths[6]

	sort.Ints(avgs)

	ranked := make([]string, 0)
	ranked = append(ranked, avgMap[avgs[0]])
	ranked = append(ranked, avgMap[avgs[1]])
	ranked = append(ranked, avgMap[avgs[2]])

	e := json.NewEncoder(w)
	e.Encode(ranked)
}

//3.2b
//rgbColourDistanceHandler returns, for a given location,
//the paths to the three most recent images, ranked in
//increasing order by the distance to a given target hex
//encoded rgb colour.
func rgbColourDistanceHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	type RequestVals struct {
		Address string
		Color   string
	}
	reqvals := RequestVals{req.FormValue("Address"), req.FormValue("Color")}

	if reqvals.Address == "" || reqvals.Color == "" {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(req)
	results, err := services.GetAddress(reqvals.Address, &ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	images, err := services.GetImages(results.Geometry.Location.Lng, results.Geometry.Location.Lat, &ctx)
	if err != nil {
		http.Error(w, "Images not found", http.StatusNotFound)
		return
	}
	imagePaths := utils.GetImagePaths(images)
	rlevel := -1

	//convert hex to rgb
	r, err := strconv.ParseUint(reqvals.Color[0:2], 16, 32)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	g, err := strconv.ParseUint(reqvals.Color[2:4], 16, 32)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	b, err := strconv.ParseUint(reqvals.Color[2:4], 16, 32)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	type Result struct {
		Message int
		Error   error
	}
	rgb := []int{int(r), int(g), int(b)}
	outs := make([][]chan Result, 0)
	outs2 := make([]chan Result, 0)
	m := 0
	avgs := make([][]int, 3)
	colourDifMap := make(map[int]string)
	ctx1 := appengine.NewContext(req)
	//for each image colour band path, create a channel
	//add it to an array specific to that image.
	for i := 0; i < len(imagePaths); i++ {
		//a channel for all bands of the image has
		//been added. Append the array and create a one
		//to hold the channels of the next image
		if i > 0 && i%3 == 0 {
			outs = append(outs, outs2)
			m++
			outs2 = make([]chan Result, 0)
		}
		out := make(chan Result)
		outs2 = append(outs2, out)
		if i == len(imagePaths)-1 {
			outs = append(outs, outs2)
		}
		path := imagePaths[i]
		//use go func to make concurrent calls to the
		//jpeg2000 analysis api, and then the avg function
		//returning the result to the appropriate channel
		go func(string, int, chan Result) {
			colour, err := services.GetColour(path, rlevel, &ctx1)
			if err != nil {
				colour, err = services.GetColour(path, rlevel, &ctx1)
				if err != nil {
					//http.Error(w, "Bad response from http://35.227.24.82/api/jp2", http.StatusBadGateway)
					out <- Result{0, err}
				}
			}
			out <- Result{utils.Avg(colour), nil}
		}(path, rlevel, out)
	}
	//the values returned by the go routines are
	//channeled to the avgs array
	for i := range outs {
		for k := range outs[i] {
			result := <-outs[i][k]
			if result.Error != nil {
				http.Error(w, result.Error.Error(), http.StatusBadGateway)
				return
			}
			avgs[i] = append(avgs[i], result.Message)
		}
	}

	//calculate the euclideanDistance for each image
	colourDifs := make([]int, 0)
	colourDifs = append(colourDifs, utils.EuclideanDistance(avgs[0], rgb))
	colourDifs = append(colourDifs, utils.EuclideanDistance(avgs[1], rgb))
	colourDifs = append(colourDifs, utils.EuclideanDistance(avgs[2], rgb))

	colourDifMap[colourDifs[0]] = imagePaths[0]
	colourDifMap[colourDifs[1]] = imagePaths[3]
	colourDifMap[colourDifs[2]] = imagePaths[6]

	sort.Ints(colourDifs)

	ranked := make([]string, 0)
	ranked = append(ranked, strings.SplitAfter(colourDifMap[colourDifs[0]], "SAFE")[0])
	ranked = append(ranked, strings.SplitAfter(colourDifMap[colourDifs[1]], "SAFE")[0])
	ranked = append(ranked, strings.SplitAfter(colourDifMap[colourDifs[2]], "SAFE")[0])

	e := json.NewEncoder(w)
	e.Encode(ranked)
}

//4.1
//For two given sets of coordinates, returns the paths of
//the blue, red and green of the three most recent images.
func moreCoordinatesHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	type RequestVals struct {
		WestLon  float64
		EastLon  float64
		NorthLat float64
		SouthLat float64
	}
	westlon, err := strconv.ParseFloat(req.FormValue("WestLon"), 64)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	eastlon, err := strconv.ParseFloat(req.FormValue("EastLon"), 64)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	northlat, err := strconv.ParseFloat(req.FormValue("NorthLat"), 64)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	southlat, err := strconv.ParseFloat(req.FormValue("SouthLat"), 64)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	reqvals := RequestVals{westlon, eastlon, northlat, southlat}
	ctx := appengine.NewContext(req)
	images, err := services.GetMoreImages(reqvals.WestLon, reqvals.EastLon, reqvals.NorthLat, reqvals.SouthLat, &ctx)
	if err != nil {
		http.Error(w, "Images not found", http.StatusNotFound)
		return
	}
	imagePaths := utils.GetImagePaths(images)
	e := json.NewEncoder(w)
	e.Encode(imagePaths)
}
