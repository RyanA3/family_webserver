package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
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

func RenameAndMove(filepath string, meta ImageMeta) {
	//In case the target directory is incorrect, avoid destroying everything in it
	if (meta.extension != "png" && meta.extension != "jpg" && meta.extension != "jpeg") || (meta.extension == "" && meta.camera_make == "") {
		fmt.Println("FATAL: Error moving file, is the directory correct?")
		return
	}

	newpath := images_dir + "/" + meta.id.Hex()
	if len(meta.extension) > 0 {
		newpath += "." + meta.extension
	}
	os.Rename(filepath, newpath)

}

func CreateMinified(meta ImageMeta) {

	//Paths of source and destination files
	src := images_dir + "/" + meta.id.Hex()
	dest := mini_images_dir + "/" + meta.id.Hex()
	if len(meta.extension) > 0 {
		src += "." + meta.extension
		dest += "." + meta.extension
	}

	//Check target file is valid
	srcFileStat, err := os.Stat(src)
	if err != nil {
		fmt.Println("FATAL: Error copying file for minification")
		return
	}

	if !srcFileStat.Mode().IsRegular() {
		fmt.Println("ERROR: Failed to copy file for minification, specified source path was invalid")
		return
	}

	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println("ERROR: Failed to open file for copying")
		return
	}
	defer srcFile.Close()

	//Decode image
	img, format, err := image.Decode(srcFile)
	if err != nil {
		fmt.Println("ERROR: Decoding image for minification")
		return
	}

	//Resize to 400 width, and keep aspect ratio
	small := resize.Resize(400, 0, img, resize.Lanczos3)

	//Create and open destination file
	destFile, err := os.Create(dest)
	if err != nil {
		fmt.Println("ERROR: Failed to create new image file to copy into")
		return
	}
	defer destFile.Close()

	//Encode resulting image data into destination file
	switch format {
	case "jpeg", "jpg":
		jpeg.Encode(destFile, small, nil)
		break
	case "png":
		png.Encode(destFile, small)
		break
	case "gif":
		gif.Encode(destFile, small, nil)
		break
	default:
		fmt.Println("ERROR: Failed to encode smaller image file, invalid or unsupported format?")
		break
	}

	fmt.Println("Finishied creating small file version")

}
