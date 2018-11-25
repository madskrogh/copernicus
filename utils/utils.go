package utils

import (
	"math"
	"strconv"
	"strings"
)

//avg takes an array of array of ints
//and the returns the average value
func Avg(data *[][]int) int {
	sum := 0
	count := 0
	numbers := *data
	for i := range numbers {
		for k := range numbers[i] {
			sum += numbers[i][k]
			count++
		}
	}
	return sum / count
}

//euclideanDistance takes two arrays of rgb colour
//values and returns the eucludian distance between
//them (https://en.wikipedia.org/wiki/Color_difference#Euclidean)
func EuclideanDistance(rgb []int, rgb1 []int) int {
	dist := math.Sqrt(math.Pow((float64(rgb[0])-float64(rgb[2])), 2) + math.Pow((float64(rgb[1])-float64(rgb[1])), 2) + math.Pow((float64(rgb[2])-float64(rgb[0])), 2))
	return int(dist)
}

//getImagePaths takes the partial image path and
//metadata retrieved from big query (see getImages)
//and constructs/returns the absolute image paths
//for the red blue and green bands according to the
//ESA SAFE standard (http://earth.esa.int/SAFE/)
func GetImagePaths(images [][]string) []string {
	var imagePaths []string
	for _, i := range images {
		one := "console.cloud.google.com/storage/browser"
		two := strings.Split(i[11], "//")[1]
		three := "GRANULE"
		four := "L1C"
		five := "T" + strings.Split(i[11], "/")[4]
		six := strings.Split(i[11], "/")[5]
		seven := strings.Split(i[11], "/")[6]
		eight := strings.Split(i[0], "_")[2]
		nine := strings.Split(i[0], "_")[3]
		ten := "IMG_DATA"
		eleven := strings.Split(i[11], "_")[2]

		for k := 2; k < 5; k++ {
			twelve := "B0" + strconv.Itoa(k) + ".jp2"
			path := one + "/" + two + "/" + three + "/" + four + "_" + five + six + seven + "_" + eight + "_" + nine + "/" + ten + "/" + five + six + seven + "_" + eleven + "_" + twelve
			imagePaths = append(imagePaths, path)
		}
	}
	return imagePaths
}
