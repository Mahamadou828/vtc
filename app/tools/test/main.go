package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

const (
	EventFilePath = "./event.local.json"
)

func main() {
	//parsing endpointURL parameter and extracting path and query parameter
	endpointURL := flag.String("endpointURL", "", "URL of the endpoint to test")
	eventFile := flag.String("eventFile", "", "URL of the endpoint to test")
	flag.Parse()

	u, err := url.Parse(*endpointURL)
	if err != nil {
		log.Fatalf("failed to parse endpointURL parameter: %v", err)
	}

	//reading the request file and extract the content in string
	log.Println("opening the request file.")
	content, err := os.ReadFile(*eventFile)
	if err != nil {
		log.Fatalf("failed to open the request file: %v.", err)
	}
	log.Println("extract content from request file.")
	data := string(content)

	//opening event file and modifying the file content
	var event events.APIGatewayProxyRequest

	log.Println("opening the event file.")
	file, err := os.Open(EventFilePath)
	if err != nil {
		log.Fatalf("failed to open the event local file: %v.", err)
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&event); err != nil {
		log.Fatalf("failed to parse the event local file: %v.", err)
	}

	//modifying event content
	log.Println("modifying the event body.")
	event.Body = data
	event.Path = u.Path

	for key, val := range u.Query() {
		//@todo handle array value
		event.QueryStringParameters[key] = val[0]
	}

	// Marshal the modified struct back to JSON
	marshalEvent, err := json.MarshalIndent(event, "", "")
	if err != nil {
		log.Fatalf("failed to marshal the modify event: %v.", err)
	}

	// Save the modified content back to the file
	log.Println("saving new event to file.")
	if err := os.WriteFile(EventFilePath, marshalEvent, 0644); err != nil {
		log.Fatalf("failed to save the modify event struct: %v.", err)
	}

	log.Println("event content successfully modified and saved.")
}
