package main

import (
	"bytes"
	"context"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type esStruct struct {
	hosts   string
	msg     string
	date    string
	esIndex string
	patient string
	doctor  string
	sentBy  string
	client  *elasticsearch.Client
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

	b.WriteString(`{"patient": "` + es.patient +
		`", "doctor": "` + es.doctor +
		`", "msg": "` + es.msg +
		`", "sentBy": "` + es.sentBy +
		`", "date": "` + es.date +
		`"}`)

	// Set up the request object.
	req := esapi.IndexRequest{
		Index: es.esIndex,
		Body:  strings.NewReader(b.String()),
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), es.client)
	if res.StatusCode != 200 || err != nil {
		log.Println(res)
		return err
	}

	defer res.Body.Close()

	return nil
}

func search(es *esStruct, buf bytes.Buffer) (*esapi.Response, error) {
	// Perform the search request.
	return es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex(es.esIndex),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
		es.client.Search.WithPretty(),
	)
}
