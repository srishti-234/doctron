package main

import (
	"context"
	"github.com/lampnick/doctron-client-go"
	"log"
)

const domain = "http://localhost:8080"
const defaultUsername = "doctron"
const defaultPassword = "lampnick"

func main() {
	client := doctron.NewClient(context.Background(), domain, defaultUsername, defaultPassword)
	req := doctron.NewDefaultHTML2ImageRequestDTO()
	req.ConvertURL = "http://doctron.lampnick.com/doctron.html"
	req.UploadKey = "test.png"
	response, err := client.HTML2ImageAndUpload(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response)
	//log.Println(len(response.Data))
}
