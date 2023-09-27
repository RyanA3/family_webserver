package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mclient *mongo.Client
var db *mongo.Database
var images_collection *mongo.Collection

func ConnectDatabase() {
	fmt.Println("Connecting database...")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(viper.GetString("MONGO_AUTH_URL")).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Printf("Connected Database\n")

	mclient = client
	db = mclient.Database(database_name)
	images_collection = db.Collection("ImageMeta")
}

func DisconnectDatabase() {

	mclient.Disconnect(context.TODO())

}

func UploadImageData(imgdat ImageMeta) {

	fmt.Printf("Uploading %s...\n", imgdat.original_name)

	res, err := images_collection.InsertOne(context.TODO(), ImageMetaToBson(imgdat))

	if err != nil {
		fmt.Printf("Error uploading document %s\n %s\n", imgdat.original_name, err)
	}
	fmt.Printf("Uploaded document with id %v\n", res.InsertedID)

}

func UpdateDuplicates(imgdat ImageMeta) {

	fmt.Printf("Finding duplicates of %s...\n", imgdat.original_name)

	var time_low = imgdat.created.Add(time.Duration(-duplicate_time_range))
	var time_high = imgdat.created.Add(time.Duration(duplicate_time_range))

	res, err := images_collection.Find(context.TODO(), bson.M{
		"_id": bson.M{
			"$ne": imgdat.id,
		},
		"camera_make":  imgdat.camera_make,
		"camera_model": imgdat.camera_model,
		"$or": []interface{}{bson.D{
			{Key: "$and", Value: []interface{}{bson.D{
				{Key: "created", Value: imgdat.created},
				{Key: "file_size", Value: imgdat.file_size},
			}}},
			{Key: "created", Value: bson.M{
				"$gte": time_low,
				"$lte": time_high,
			}},
		}},
	},
	)

	defer res.Close(context.TODO())

	if err != nil {
		fmt.Printf("Error whilst querying: %s\n", err)
		return
	}

	var duplicates []ImageMeta
	// res.All(context.TODO(), duplicates)

	for res.Next(context.TODO()) {
		var result interface{}
		err = res.Decode(&result)

		if err != nil {
			fmt.Printf("Error whilst decoding: %s\n", err)
			continue
		}
		//fmt.Printf("result: %v\n", result)

		meta := DecodeImageMetaFromDatabaseResponse(result.(primitive.D))
		duplicates = append(duplicates, meta)
		//fmt.Printf("name: %s\nsize: %v\ncreated: %v\nmake: %s\nmodel: %s\n\n", meta.original_name, meta.file_size, meta.created, meta.camera_make, meta.camera_model)
		//fmt.Printf("name: %s\nsize: %v\ncreated: %v\nmake: %s\nmodel: %s\n\n", result.original_name, result.file_size, result.created, result.camera_make, result.camera_model)

		//duplicates = append(duplicates, result)
	}

	fmt.Printf("Found %s duplicates: \n", fmt.Sprint(len(duplicates)))
	for _, dupe := range duplicates {
		fmt.Printf("name: %s\nsize: %v\ncreated: %v\nmake: %s\nmodel: %s\n\n", dupe.original_name, dupe.file_size, dupe.created, dupe.camera_make, dupe.camera_model)
	}

}

func DecodeImageMetaFromDatabaseResponse(data primitive.D) ImageMeta {

	var out ImageMeta

	for _, v := range data {
		switch v.Key {
		case "_id":
			out.id = v.Value.(primitive.ObjectID)
		case "original_name":
			out.original_name = v.Value.(string)
		case "extension":
			out.extension = v.Value.(string)
		case "created":
			out.created = v.Value.(primitive.DateTime).Time()
		case "uploaded":
			out.uploaded = v.Value.(primitive.DateTime).Time()
		case "camera_make":
			out.camera_make = v.Value.(string)
		case "camera_model":
			out.camera_model = v.Value.(string)
		case "file_size":
			out.file_size = v.Value.(int64)
		case "duplicates":
			if v.Value != nil {
				out.duplicates = v.Value.([]string)
			}
		default:
			fmt.Printf("Its jover -- Field DIDNT Load {%s : some_value}", v.Key)
		}
	}

	// for i := 0; i < element.NumField(); i++ {
	// 	field := element.Field(i)
	// 	field_type := element.Field(i).Name

	// 	switch field_type {
	// 	case "_id":
	// 		out.id = primitive.ObjectID{byte(field.Type.Elem().Bits())}
	// 	case "original_name":
	// 		out.original_name = field.Type.Elem().String()
	// 	default:
	// 		fmt.Printf("F\n")
	// 	}
	// }

	return out

}
