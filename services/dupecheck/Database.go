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
	images_collection = db.Collection(images_collection_name)

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

func UpdateDuplicates(imgdat ImageMeta) int {

	//fmt.Printf("Finding duplicates of %s...\n", imgdat.original_name)

	//Find the duplicates of this image
	var time_low = imgdat.created.Add(time.Duration(-duplicate_time_range))
	var time_high = imgdat.created.Add(time.Duration(duplicate_time_range))

	opts := options.Find().SetProjection(bson.M{
		"_id":           1,
		"camera_make":   0,
		"camera_model":  0,
		"original_name": 0,
		"file_size":     0,
		"duplicates":    0,
		"created":       0,
		"uploaded":      0,
		"extension":     0,
	})

	query := bson.M{
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
	}

	cursor, err := images_collection.Find(context.TODO(), query, opts)
	defer cursor.Close(context.TODO())

	if err != nil {
		fmt.Printf("Error whilst querying: %s\n", err)
		return 0
	}

	duplicates := DecodeObjectIds(cursor)

	//Update the duplicates of the document
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "duplicates", Value: duplicates},
		}},
	}

	_, err = images_collection.UpdateByID(context.TODO(), imgdat.id, update)

	if err != nil {
		fmt.Printf("Error updating duplicates for %s:\n%s", imgdat.id.Hex(), err)
		return 0
	}

	return len(duplicates)

}

func DecodeObjectIds(cursor *mongo.Cursor) []primitive.ObjectID {

	var out []primitive.ObjectID

	for cursor.Next(context.TODO()) {
		var result interface{}
		cursor.Decode(&result)
		var data = result.(primitive.D)[0]
		out = append(out, data.Value.(primitive.ObjectID))
	}

	return out

}

func DecodeImageMetas(cursor *mongo.Cursor) []ImageMeta {

	var out []ImageMeta

	for cursor.Next(context.TODO()) {
		var result interface{}
		cursor.Decode(&result)
		out = append(out, DecodeImageMeta(result.(primitive.D)))
	}

	return out

}

// TODO: There has to be a better way to do this, custom decoder if feeling funny, library if not feeling it
func DecodeImageMeta(data primitive.D) ImageMeta {

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
