package main

import (
	"context"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type esStruct struct {
	hosts       string
	msg         string
	esIndex     string
	chatClients string
	client      *elasticsearch.Client
}

func connectES(es *esStruct) (*elasticsearch.Client, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			es.hosts,
		},
	}

	return elasticsearch.NewClient(cfg)
}

func index(es *esStruct) error {

	// Build the request body.
	var b strings.Builder
	b.WriteString(`{"client" : "` + es.chatClients + `", "msg": "` + es.msg + `"}`)

	// Set up the request object.
	req := esapi.IndexRequest{
		Index: es.esIndex,
		Body:  strings.NewReader(b.String()),
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
