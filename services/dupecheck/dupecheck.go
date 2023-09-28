package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"

	"github.com/spf13/viper"

	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: Run go function from node https://medium.com/learning-the-go-programming-language/calling-go-functions-from-other-languages-4c7d8bcc69bf
// ref hasn't been updated in 6 years and doesn't work
// node-ffi only supports up to node 11 :(
// TODO: Find another way to run go functions from node
// Webassembly????? https://pedromarquez.dev/blog/2023/2/node_golang_wasm
// TODO: Call go functions from node server using compile to c with go and a node library?
// TODO: Connect / pass mongoose to go program from node if possible. Add image metadata to database on process
// TODO: Setup local mongo database for the ungodly amount of database calls that are about to happen

var process_dir = "/home/ryan/repos/family_image_server/files/processing"
var images_dir = "/home/ryan/repos/family_image_server/files/images"
var output_dir = "/home/ryan/repos/family_image_server/files/processing.json"
var database_name = "FamilyDB"

const env_path = "/home/ryan/repos/family_image_server/.env"

var default_creation_time = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local)

func initConstants() {
	viper.GetViper().SetConfigFile(env_path)
	viper.ReadInConfig()

	files_dir := viper.GetString("FILES_DIR")
	process_dir = files_dir + viper.GetString("PROCESSING_FOLDER")
	images_dir = files_dir + viper.GetString("IMAGES_FOLDER")
	output_dir = files_dir + viper.GetString("DUPECHECK_OUTPUT_FILE")
	database_name = viper.GetString("MONGO_DATABASE")

	fmt.Printf("Loaded env with viper")
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
func decode(filepath string) ImageMeta {

	//Get the file extension
	splitpath := strings.Split(filepath, ".")
	var extension string = ""
	if len(splitpath) > 1 {
		extension = splitpath[len(splitpath)-1]
	}

	var meta ImageMeta = ImageMeta{id: primitive.NewObjectID(), extension: extension, original_name: filepath, uploaded: time.Now()}

	//Move the file to the images folder once processing has complete, and rename it
	defer func() {
		newpath := images_dir + "/" + meta.id.Hex()
		if len(meta.extension) > 0 {
			newpath += "." + meta.extension
		}
		os.Rename(filepath, newpath)
	}()

	//Read metadata
	filemeta, err := os.Stat(filepath)
	if err != nil {
		fmt.Println("Error reading image meta: " + filepath)
		meta.created = default_creation_time
		return meta
	}

	meta.file_size = filemeta.Size()
	meta.original_name = filemeta.Name()

	//Open file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error loading image file: " + filepath)
		return meta
	}

	defer file.Close()

	//Decode Exif data
	exif.RegisterParsers(mknote.Canon, mknote.NikonV3)
	xdata, err := exif.Decode(file)
	if err != nil {
		file.Close()
		fmt.Println("Error reading image exif data: " + filepath)
		return meta
	}

	// Convenience functions for getting datetime
	meta.created, err = xdata.DateTime()

	//Try to get make/model as well
	make, err := xdata.Get(exif.Make)
	if err == nil {
		meta.camera_make, err = make.StringVal()
	}
	model, err := xdata.Get(exif.Model)
	if err == nil {
		meta.camera_model, err = model.StringVal()
	}

	return meta
}

/*
Takes
  - A root directory
  - A list of files to decode

Returns
  - A list of ImageExifData (decoded images)
*/
func decodeAll(path string, files []fs.DirEntry) []ImageMeta {

	var xdatas = make([]ImageMeta, len(files))
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

	var imgdats []ImageMeta = decodeAll(process_dir, files)

	// for _, imgdat := range imgdats {
	// 	fmt.Printf("%s has %s duplicates: ", imgdat.original_name, fmt.Sprint(len(imgdat.duplicates)))

	// 	if len(imgdat.duplicates) > 0 {
	// 		for _, dupe := range imgdat.duplicates {
	// 			fmt.Printf("%s, ", dupe)
	// 		}
	// 	}

	// 	fmt.Println()
	// }

	uploadImagesDataToDatabase(imgdats)
	UpdateDuplicates(imgdats[0])

}

func uploadImagesDataToDatabase(imgdats []ImageMeta) {
	for _, imgdat := range imgdats {
		UploadImageData(imgdat)
	}
}

/*
Takes
  - A list of images currently in the processing folder

Returns
  - Any duplicates of those images within the same folder
*/
func findDuplicatesInProcessing(imgdats []ImageMeta) {
	for i, f := range imgdats {
		f.duplicates = findDuplicatesOf(f, imgdats, i+1)
	}
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
func findDuplicatesOf(file ImageMeta, check []ImageMeta, start_index int) []string {
	//fmt.Printf("Checking %s against: ", file.original_name)
	//fmt.Println()

	var duplicate_of []string
	for i := start_index; i < len(check); i++ {
		if isDuplicate(file, check[i]) {
			duplicate_of = append(duplicate_of, check[i].id.String())
		}
	}

	return duplicate_of
}

// Photos taken within 2 seconds of each other considered duplicates if taken on the same device
// Photos which don't have a date taken (their date is set to default time) will be ignored for this check
var duplicate_time_range = time.Duration.Seconds(2)

func isDuplicate(file, check ImageMeta) bool {
	//fmt.Printf("%s: %s // %s: %s", file.original_name, file.created, check.original_name, check.created)
	//fmt.Println()
	if file.camera_make != check.camera_make {
		return false
	}
	if file.camera_model != check.camera_model {
		return false
	}

	if file.file_size == check.file_size &&
		file.created == check.created {
		return true
	}

	if check.created != default_creation_time &&
		check.created.After(file.created.Add(time.Duration(-duplicate_time_range))) &&
		check.created.Before(file.created.Add(time.Duration(duplicate_time_range))) {
		return true
	}

	return false
}

func main() {

	initConstants()

	ConnectDatabase()
	defer DisconnectDatabase()

	//fmt.Println("Checking for duplicates!")
	//ProcessUploadedImages(process_dir, images_dir)

	//1. Decode and load all files in processing

	//2. Goroutine upload files to database and move to images folder

	//3. Wait to completion

	//4. Goroutine check files for duplicates and update

	//5. Done

}
