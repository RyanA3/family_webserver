package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
)

// TODO: Read metadata from image file https://stackoverflow.com/questions/60497938/read-exif-metadata-with-go
// TODO: Call go functions from node server using compile to c with go and a node library?
// TODO: Connect / pass mongoose to go program from node if possible. Add image metadata to database on process
// TODO: Setup local mongo database for the ungodly amount of database calls that are about to happen

const process_dir = "/home/ryan/repos/family_image_server/files/processing"
const success_dir = "/home/ryan/repos/family_image_server/files/images"

func ProcessUploadedImages(process_dir, success_dir string) {
	//Load each file in processing
	files, err := os.ReadDir(process_dir)
	num_files := len(files)
	if err != nil {
		log.Fatal(err)
	}

	//Load metadata of each file in processing
	var filemetas []os.FileInfo
	for _, f := range files {
		filemeta, err := os.Stat(process_dir + "/" + f.Name())
		if err != nil {
			fmt.Printf("Error loading meta for file %s", f.Name())
			continue
		}

		filemetas = append(filemetas, filemeta)
	}

	//Check for duplicates
	fmt.Printf("Found %s files\n", fmt.Sprint(num_files))
	for i, f := range filemetas {
		//fmt.Println(fmt.Sprint(i) + " " + f.Name())
		findDuplicates(f, filemetas, i+1)
	}
}

func findDuplicates(file fs.FileInfo, check []fs.FileInfo, start_index int) []fs.FileInfo {
	fmt.Printf("Checking %s against: ", file.Name())
	fmt.Println()

	var duplicate_of []fs.FileInfo
	for i := start_index; i < len(check); i++ {
		if isDuplicate(file, check[i]) {
			duplicate_of = append(duplicate_of, check[i])
		}
	}

	return duplicate_of
}

func isDuplicate(file, check fs.FileInfo) bool {
	fmt.Printf("%s: %s // %s: %s", file.Name(), file.ModTime().String(), check.Name(), check.ModTime())
	fmt.Println()

	return false
}

func main() {
	fmt.Println("Main shouldn't be used!!!!")
	ProcessUploadedImages(process_dir, success_dir)
}
