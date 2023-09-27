package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"C"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"

	"github.com/spf13/viper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
import (
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

var mclient *mongo.Client
var db *mongo.Database

type ImageExifData struct {
	created       time.Time
	make          string
	model         string
	original_name string
	file_size     int64
}

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

func connectDatabase() {

	fmt.Println("Connecting database...")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(viper.GetString("MONGO_AUTH_URL")).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		fmt.Printf("Disconnected db %s\n", err)
	// 		panic(err)
	// 	}
	// }()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Printf("Connected Database\n")

	mclient = client
	db = mclient.Database(database_name)

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

	var meta ImageMeta = ImageMeta{id: primitive.NewObjectID(), original_name: filepath, uploaded: time.Now()}

	//Read metadata
	filemeta, err := os.Stat(filepath)
	if err != nil {
		fmt.Println("Error reading image meta: " + filepath)
		meta.created = default_creation_time
		return meta
		//return ImageExifData{default_time, default_make, default_model, time.Now().String(), default_size}
	}

	meta.file_size = uint64(filemeta.Size())
	meta.uploaded = filemeta.ModTime()
	meta.original_name = filemeta.Name()

	//Open file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error loading image file: " + filepath)
		return meta
		//return ImageExifData{default_time, default_make, default_model, filemeta.Name(), filemeta.Size()}
	}

	//Decode Exif data
	exif.RegisterParsers(mknote.Canon, mknote.NikonV3)
	xdata, err := exif.Decode(file)
	if err != nil {
		file.Close()
		fmt.Println("Error reading image exif data: " + filepath)
		return meta
		//return ImageExifData{default_time, default_make, default_model, filemeta.Name(), filemeta.Size()}
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

	file.Close()
	return meta
	//return ImageExifData{time, smake, smodel, filemeta.Name(), filemeta.Size()}

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

	//os.OpenFile(output_dir, os.O_CREATE, fs.FileMode(os.O_RDWR))
	uploadImagesDataToDatabase(imgdats, dupes)

}

func uploadImagesDataToDatabase(imgdats []ImageMeta, dupes [][]string) {

	images_collection := db.Collection("ImageMeta")
	images_collection.DeleteMany(context.TODO(), &ImageMeta{})

	for i, imgdat := range imgdats {
		uploadImage(images_collection, imgdat, dupes[i])
	}

}

type ImageMeta struct {
	id            primitive.ObjectID `bson:"_id"`
	original_name string             `bson:"original_name"`
	created       time.Time          `bson:"created"`
	uploaded      time.Time          `bson:"uploaded"`
	file_size     uint64             `bson:"file_size"`
	camera_make   string             `bson:"camera_make,omitempty"`
	camera_model  string             `bson:"camera_model,omitempty"`
	duplicates    []string           `bson:"duplicates,omitempty"`
}

func ImageMetaToBson(meta ImageMeta) primitive.D {
	return bson.D{
		{Key: "_id", Value: meta.id},
		{Key: "original_name", Value: meta.original_name},
		{Key: "camera_make", Value: meta.camera_make},
		{Key: "camera_model", Value: meta.camera_model},
		{Key: "created", Value: meta.created},
		{Key: "uploaded", Value: meta.uploaded},
		{Key: "file_size", Value: meta.file_size},
		{Key: "duplicates", Value: meta.duplicates},
	}
}

func uploadImage(images_collection *mongo.Collection, imgdat ImageMeta, dupes []string) {

	fmt.Printf("Uploading %s...\n", imgdat.original_name)

	// res, err := images_collection.InsertOne(context.Background(), bson.D{
	// 	{Key: "original_name", Value: imgdat.original_name},
	// 	{Key: "camera_make", Value: imgdat.make},
	// 	{Key: "camera_model", Value: imgdat.model},
	// 	{Key: "created", Value: imgdat.created},
	// 	{Key: "uploaded", Value: time.Now()},
	// 	{Key: "file_size", Value: imgdat.file_size},
	// 	{Key: "duplicates", Value: dupes},
	// })
	imgdat.duplicates = dupes
	res, err := images_collection.InsertOne(context.TODO(), ImageMetaToBson(imgdat))

	if err != nil {
		fmt.Printf("Error uploading document %s\n %s\n", imgdat.original_name, err)
	}

	fmt.Printf("Uploaded document with id %v\n", res.InsertedID)
}

/*
Takes
  - A list of images currently in the processing folder

Returns
  - Any duplicates of those images within the same folder
*/
func findDuplicatesInProcessing(imgdats []ImageMeta) [][]string {
	var dupes [][]string = make([][]string, len(imgdats))

	for i, f := range imgdats {
		dupes[i] = findDuplicatesOf(f, imgdats, i+1)
	}

	return dupes
}

/*
Checks if each image is a duplicate of an image that has already been uploaded and is in the database (not in processing)
Takes
  - A list of images to check

Returns
  - A list of duplicates from the database corresponding to each image in the input list
*/
// func findDuplicatesInDatabase(imgdats []ImageExifData) [][]string {

// }

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
	connectDatabase()
	fmt.Println("Checking for duplicates!")
	ProcessUploadedImages(process_dir, images_dir)
	mclient.Disconnect(context.TODO())
}
