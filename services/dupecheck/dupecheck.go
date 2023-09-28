package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
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

func ProcessUploadedImages(process_dir, success_dir string) {
	//1. Load and decode all files in processing folder
	files, err := os.ReadDir(process_dir)
	num_files := len(files)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %s files\n", fmt.Sprint(num_files))
	var imgdats []ImageMeta = decodeAll(process_dir, files)

	//2. Goroutine upload files to database and move to images folder
	uploadImagesDataToDatabase(imgdats)

	//3. Wait to completion

	//4. Goroutine check files for duplicates and update
	UpdateDuplicates(imgdats[0])

	//5. Done

}

func uploadImagesDataToDatabase(imgdats []ImageMeta) {
	for _, imgdat := range imgdats {
		UploadImageData(imgdat)
	}
}

func main() {

	initConstants()

	ConnectDatabase()
	defer DisconnectDatabase()

	//fmt.Println("Checking for duplicates!")
	//ProcessUploadedImages(process_dir, images_dir)

}
