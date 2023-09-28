package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
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
var duplicate_time_range time.Duration = time.Duration(2 * float64(time.Second))

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

func ProcessUploadedImages(process_dir, success_dir string) {
	//1. Load all files in processing folder
	files, err := os.ReadDir(process_dir)
	num_files := len(files)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %s files\n\nStarting decode and upload process...\n", fmt.Sprint(num_files))

	var start_decoding_time = time.Now()

	var uploaded []ImageMeta = make([]ImageMeta, len(files))
	channel := make(chan ImageMeta, len(files))

	//2. Handle all the files in processing
	for i, file := range files {
		go func(file fs.DirEntry, i int, c chan ImageMeta) {
			filepath := process_dir + "/" + file.Name()
			fmt.Printf("Decoding file %s\n", filepath)
			//2.1 Decode file
			meta := decode(filepath)

			//2.2 Upload to database
			UploadImageData(meta)

			//2.3 Move to images folder
			RenameAndMove(filepath, meta)

			c <- meta

		}(file, i, channel)
	}

	//3. Receive data from decoding process and wait to completion
	for i := range files {
		uploaded[i] = <-channel
		fmt.Printf("Processed image: %v\n", uploaded[i].original_name)
	}

	var end_decoding_time = time.Now()
	var ms_taken = end_decoding_time.Sub(start_decoding_time).Milliseconds()
	fmt.Printf("%s images decoded and uploaded in %sms\n\n", fmt.Sprint(len(files)), fmt.Sprint(ms_taken))

	//4. Goroutine check files for duplicates and update
	fmt.Printf("\n\nChecking for duplicates...\n")
	var start_dupecheck_time = time.Now()

	var num_duplicates_channel = make(chan int, len(uploaded))

	for _, imgdat := range uploaded {
		go func(imgdat ImageMeta, c chan int) {
			c <- UpdateDuplicates(imgdat)
		}(imgdat, num_duplicates_channel)
	}

	//Just blocking until all duplicate update checks are done for funzies
	num_duplicates := 0
	for i := range uploaded {
		num_duplicates += <-num_duplicates_channel
		if i >= len(uploaded) {
			break
		}
	}

	var end_dupecheck_time = time.Now()
	ms_taken = end_dupecheck_time.Sub(start_dupecheck_time).Milliseconds()
	fmt.Printf("%s duplicates identified in %sms\n\n", fmt.Sprint(num_duplicates), fmt.Sprint(ms_taken))

}

func ProcessUploadedImagesSerially(process_dir, success_dir string) {
	//1. Load all files in processing folder
	files, err := os.ReadDir(process_dir)
	num_files := len(files)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %s files\n\nStarting SERIAL decode and upload process...\n", fmt.Sprint(num_files))
	var start_decoding_time = time.Now()
	var uploaded []ImageMeta = make([]ImageMeta, len(files))

	//2. Handle all the files in processing
	for i, file := range files {
		filepath := process_dir + "/" + file.Name()
		fmt.Printf("Decoding file %s\n", filepath)
		//2.1 Decode file
		uploaded[i] = decode(filepath)

		//2.2 Upload to database
		UploadImageData(uploaded[i])

		//2.3 Move to images folder
		RenameAndMove(filepath, uploaded[i])

	}

	var end_decoding_time = time.Now()
	var ms_taken = end_decoding_time.Sub(start_decoding_time).Milliseconds()
	fmt.Printf("SERIAL %s images decoded and uploaded in %sms\n\n", fmt.Sprint(len(files)), fmt.Sprint(ms_taken))

	//4. Check files for duplicates and update
	fmt.Printf("\n\nChecking for duplicates...\n")
	var start_dupecheck_time = time.Now()

	var num_duplicates = 0

	for _, imgdat := range uploaded {
		num_duplicates += UpdateDuplicates(imgdat)
	}

	var end_dupecheck_time = time.Now()
	ms_taken = end_dupecheck_time.Sub(start_dupecheck_time).Milliseconds()
	fmt.Printf("SERIAL %s duplicates identified in %sms\n\n", fmt.Sprint(num_duplicates), fmt.Sprint(ms_taken))
}

func main() {

	initConstants()

	ConnectDatabase()
	defer DisconnectDatabase()

	images_collection.DeleteMany(context.TODO(), bson.D{})

	fmt.Println("Checking for duplicates!")
	ProcessUploadedImages(process_dir, images_dir)

}
