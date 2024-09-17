package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImageMeta struct {
	id            primitive.ObjectID `bson:"_id"`
	extension     string             `bson:"extension"`
	original_name string             `bson:"original_name"`
	created       time.Time          `bson:"created"`
	uploaded      time.Time          `bson:"uploaded"`
	file_size     int64              `bson:"file_size"`
	camera_make   string             `bson:"camera_make,omitempty"`
	camera_model  string             `bson:"camera_model,omitempty"`
	duplicates    []string           `bson:"duplicates,omitempty"`
	lat           float64            `bson:"lat,omitempty"`
	long          float64            `bson:"long,omitempty"`
}

func ImageMetaToBson(meta ImageMeta) primitive.D {
	out := bson.D{
		{Key: "_id", Value: meta.id},
		{Key: "extension", Value: meta.extension},
		{Key: "original_name", Value: meta.original_name},
		{Key: "camera_make", Value: meta.camera_make},
		{Key: "camera_model", Value: meta.camera_model},
		{Key: "created", Value: meta.created},
		{Key: "uploaded", Value: meta.uploaded},
		{Key: "file_size", Value: meta.file_size},
		{Key: "duplicates", Value: meta.duplicates},
	}

	if meta.lat != 0 && meta.long != 0 {
		out = append(out, bson.E{Key: "lat", Value: meta.lat})
		out = append(out, bson.E{Key: "long", Value: meta.long})
	}

	return out
}
