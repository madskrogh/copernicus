package services

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"strconv"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/appengine/urlfetch"
	"googlemaps.github.io/maps"
)

//getAddress connects to the Google Geocoding api and
//retrieves/return the information, including
//coordinates, for the given adress
func GetAddress(address string, ctx *context.Context) (maps.GeocodingResult, error) {
	type GeocodingRequest struct {
		Address      string
		Components   map[maps.Component]string
		Bounds       *maps.LatLngBounds
		Region       string
		LatLng       *maps.LatLng
		ResultType   []string
		LocationType []maps.GeocodeAccuracy
		PlaceID      string
		Language     string
		Custom       url.Values
	}
	client := urlfetch.Client(*ctx)
	mapClient, err := maps.NewClient(maps.WithAPIKey("API KEY HERE"), maps.WithHTTPClient(client))
	if err != nil {
		return maps.GeocodingResult{}, err
	}
	r := maps.GeocodingRequest{Address: address}
	results, err := mapClient.Geocode(*ctx, &r)
	if err != nil {
		return maps.GeocodingResult{}, err
	}
	return results[0], nil
}

//getImages connects to the Google big query api and the https://bigquery.cloud.google.com/table/bigquery-public-data:cloud_storage_geo_index.sentinel_2_index
//table and retrieves/returns metadata for the three most recent
//satelite images matching the given coordinates
func GetImages(lon float64, lat float64, ctx *context.Context) ([][]string, error) {
	creds := []byte(`JSON CREDENTIALS HERE`)

	bigqueryClient, err := bigquery.NewClient(*ctx, "APP ENGINE projectID HERE", option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, errors.New("1")
	}
	q := bigqueryClient.Query(
		`SELECT * 
		FROM` + "`bigquery-public-data.cloud_storage_geo_index.sentinel_2_index`" +
			`WHERE 
		(west_lon < @target_lon AND east_lon > @target_lon)
		AND 
		(south_lat < @target_lat AND north_lat > @target_lat)
		ORDER BY
    	sensing_time DESC
		LIMIT 3;`)

	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "target_lon",
			Value: lon,
		},
		{
			Name:  "target_lat",
			Value: lat,
		},
	}
	job, err := q.Run(*ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(*ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}
	it, err := job.Read(*ctx)
	var stringRows [][]string
	for {
		var queryRow []bigquery.Value
		err := it.Next(&queryRow)
		if err == iterator.Done {
			return stringRows, nil
		}
		if err != nil {
			return nil, err
		}
		var stringRow []string
		for i := range queryRow {
			var stringElement string
			if queryRow[i] == nil {
				stringElement = ""
				stringRow = append(stringRow, stringElement)
				continue
			}
			varType := reflect.TypeOf(queryRow[i]).String()
			if varType == "float64" {
				stringElement = strconv.FormatFloat(queryRow[i].(float64), 'f', 6, 64)
			} else if varType == "int64" {
				stringElement = strconv.FormatInt(queryRow[i].(int64), 10)
			} else {
				stringElement = queryRow[i].(string)
			}
			stringRow = append(stringRow, stringElement)
		}
		stringRows = append(stringRows, stringRow)
	}
}

//getMoreImages connects to the Google big query api and the https://bigquery.cloud.google.com/table/bigquery-public-data:cloud_storage_geo_index.sentinel_2_index
//table and retrieves/returns metadata for the three most recent
//satelite images matching the given, two sets of coordinates
func GetMoreImages(west_lon float64, east_lon float64, south_lat float64, north_lat float64, ctx *context.Context) ([][]string, error) {
	creds := []byte(`JSON CREDENTIALS HERE`)

	bigqueryClient, err := bigquery.NewClient(*ctx, "APP ENGINE projectID HERE", option.WithCredentialsJSON(creds))

	if err != nil {
		return nil, err
	}
	q := bigqueryClient.Query(
		`SELECT * 
		FROM` + "`bigquery-public-data.cloud_storage_geo_index.sentinel_2_index`" +
			`WHERE 
		(west_lon < @target_west_lon AND east_lon > @target_east_lon)
		AND 
		(south_lat < @target_south_lat AND north_lat > @target_north_lat)
		ORDER BY
    	sensing_time DESC;`)

	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "target_west_lon",
			Value: west_lon,
		},
		{
			Name:  "target_east_lon",
			Value: east_lon,
		},
		{
			Name:  "target_south_lat",
			Value: south_lat,
		},
		{
			Name:  "target_north_lat",
			Value: north_lat,
		},
	}
	job, err := q.Run(*ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(*ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}
	it, err := job.Read(*ctx)
	var stringRows [][]string
	for {
		var queryRow []bigquery.Value
		err := it.Next(&queryRow)
		if err == iterator.Done {
			return stringRows, nil
		}
		if err != nil {
			return nil, err
		}
		var stringRow []string
		for i := range queryRow {
			var stringElement string
			if queryRow[i] == nil {
				stringElement = ""
				stringRow = append(stringRow, stringElement)
				continue
			}
			varType := reflect.TypeOf(queryRow[i]).String()
			if varType == "float64" {
				stringElement = strconv.FormatFloat(queryRow[i].(float64), 'f', 6, 64)
			} else if varType == "int64" {
				stringElement = strconv.FormatInt(queryRow[i].(int64), 10)
			} else {
				stringElement = queryRow[i].(string)
			}
			stringRow = append(stringRow, stringElement)
		}
		stringRows = append(stringRows, stringRow)
	}
}
