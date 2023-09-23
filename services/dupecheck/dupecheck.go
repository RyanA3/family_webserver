package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

// TODO: Read metadata from image file https://stackoverflow.com/questions/60497938/read-exif-metadata-with-go
// TODO: Call go functions from node server using compile to c with go and a node library?
// TODO: Connect / pass mongoose to go program from node if possible. Add image metadata to database on process
// TODO: Setup local mongo database for the ungodly amount of database calls that are about to happen

const process_dir = "/home/ryan/repos/family_image_server/files/processing"
const success_dir = "/home/ryan/repos/family_image_server/files/images"

var default_time = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local)
var default_make = "None"
var default_model = "None"
var default_size int64 = 0

type ImageExifData struct {
	created       time.Time
	make          string
	model         string
	original_name string
	file_size     int64
}

/*
Takes
  - a path to an image file

Returns
  - Time of creation
  - Camera Make
  - Camera Model
  - Original File name
  - File Size
*/
func decode(filepath string) ImageExifData {

	//Read metadata
	filemeta, err := os.Stat(filepath)
	if err != nil {
		fmt.Println("Error reading image meta: " + filepath)
		return ImageExifData{default_time, default_make, default_model, time.Now().String(), default_size}
	}

	//Open file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error loading image file: " + filepath)
		return ImageExifData{default_time, default_make, default_model, filemeta.Name(), filemeta.Size()}
	}

	//Decode Exif data
	exif.RegisterParsers(mknote.Canon, mknote.NikonV3)
	xdata, err := exif.Decode(file)
	if err != nil {
		file.Close()
		fmt.Println("Error reading image exif data: " + filepath)
		return ImageExifData{default_time, default_make, default_model, filemeta.Name(), filemeta.Size()}
	}

	// Convenience functions for getting datetime
	time, _ := xdata.DateTime()

	//fmt.Println("Image: " + file.Name())
	//fmt.Println("Date taken: " + time.String())

	//Try to get make/model as well
	make, make_err := xdata.Get(exif.Make)
	model, model_err := xdata.Get(exif.Model)

	var smake string
	var smodel string

	if make_err != nil {
		smake = default_make
	} else {
		smake = make.String()
	}
	if model_err != nil {
		smodel = default_model
	} else {
		smodel = model.String()
	}

	file.Close()
	return ImageExifData{time, smake, smodel, filemeta.Name(), filemeta.Size()}

}

/*
Takes
  - A root directory
  - A list of files to decode

Returns
  - A list of ImageExifData (decoded images)
*/
func decodeAll(path string, files []fs.DirEntry) []ImageExifData {

	var xdatas = make([]ImageExifData, len(files))
	for i, file := range files {
		xdatas[i] = decode(path + "/" + file.Name())
	}
	return xdatas

}

func ProcessUploadedImages(process_dir, success_dir string) {
	//Load all files in processing
	files, err := os.ReadDir(process_dir)
	num_files := len(files)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %s files\n", fmt.Sprint(num_files))

	var imgdats []ImageExifData = decodeAll(process_dir, files)
	var dupes [][]string = findDuplicatesInProcessing(imgdats)

	for i, idupes := range dupes {
		fmt.Printf("%s has %s duplicates: ", imgdats[i].original_name, fmt.Sprint(len(idupes)))

		if len(idupes) > 0 {
			for _, dupe := range idupes {
				fmt.Printf("%s, ", dupe)
			}
		}

		fmt.Println()
	}

}

/*
Takes
  - A list of images currently in the processing folder

Returns
  - Any duplicates of those images within the same folder
*/
func findDuplicatesInProcessing(imgdats []ImageExifData) [][]string {
	var dupes [][]string = make([][]string, len(imgdats))

	for i, f := range imgdats {
		dupes[i] = findDuplicatesOf(f, imgdats, i+1)
	}

	return dupes
}

/*
Generic function for finding any duplicates to an individual image in a list of images
Takes
  - A file to check against a list of files
  - A list of files to check for duplicates
  - The index to start checking files in the list
  - An index to skip (the index of the individual file, if it is contained within the list as well)

Returns
  - A list of the names of the files that duplicate the target file
*/
func findDuplicatesOf(file ImageExifData, check []ImageExifData, start_index int) []string {
	//fmt.Printf("Checking %s against: ", file.original_name)
	//fmt.Println()

	var duplicate_of []string
	for i := start_index; i < len(check); i++ {
		if isDuplicate(file, check[i]) {
			duplicate_of = append(duplicate_of, check[i].original_name)
		}
	}

	return duplicate_of
}

// Photos taken within 2 seconds of each other considered duplicates if taken on the same device
// Photos which don't have a date taken (their date is set to default time) will be ignored for this check
var duplicate_time_range = time.Duration.Seconds(2)

func isDuplicate(file, check ImageExifData) bool {
	//fmt.Printf("%s: %s // %s: %s", file.original_name, file.created, check.original_name, check.created)
	//fmt.Println()

	if file.file_size == check.file_size &&
		file.make == check.make &&
		file.model == check.model &&
		file.created == check.created {
		return true
	}

	if check.created != default_time &&
		check.created.After(file.created.Add(time.Duration(-duplicate_time_range))) &&
		check.created.Before(file.created.Add(time.Duration(duplicate_time_range))) &&
		file.make == check.make && file.model == check.model {
		return true
	}

	return false
}

func main() {
	fmt.Println("Main shouldn't be used!!!!")
	ProcessUploadedImages(process_dir, success_dir)
}
