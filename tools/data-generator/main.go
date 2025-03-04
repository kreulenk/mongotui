// This tool is a quickly thrown together data generator so create data that is used during testing, development, and
// the Demo gif in the README of this project

package main

import (
	"context"
	"fmt"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide a valid mongodb connection string")
		os.Exit(1)
	}
	connectionString := os.Args[1]
	clientOps := options.Client()
	clientOps.SetTimeout(mongoengine.Timeout)
	if !strings.Contains(connectionString, "://") {
		connectionString = "mongodb://" + connectionString
	}
	clientOps.ApplyURI(connectionString)
	client, err := mongo.Connect(clientOps)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for col := range 100 {
		docs := generateDocuments()
		_, err = client.Database("veryLargeDb").Collection(fmt.Sprintf("exampleCollection%d", col)).InsertMany(context.Background(), docs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func generateDocuments() []bson.M {
	var newDocs []bson.M
	for i := range 1000 {
		newDoc := map[string]interface{}{
			"someNumber": i,
			"someString": fmt.Sprintf("the num %d", i),
		}
		newDocs = append(newDocs, newDoc)
	}
	return newDocs
}
