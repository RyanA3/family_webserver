package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
