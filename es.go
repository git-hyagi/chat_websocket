package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

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

// connect to elasticsearch
func connectES(es *esStruct) (*elasticsearch.Client, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			es.hosts,
		},
	}

	return elasticsearch.NewClient(cfg)
}

// store data
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

// Perform the search request.
func search(es *esStruct, buf bytes.Buffer) (*esapi.Response, error) {
	return es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex(es.esIndex),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
		es.client.Search.WithPretty(),
	)
}

// Search for the chat messages between a specific doctor/patient
func retrieveMessages(lastNMsg int, patient, doctor string, es *esStruct) ([]msg, error) {
	var r map[string]interface{}
	var buf bytes.Buffer
	var esMsgs []msg

	// prepare the query
	// - bring only the last N messages (const lastNMsg)
	// - ordered by date (only the earliest messages)
	// - matching patient with doctor
	query := map[string]interface{}{
		"size": lastNMsg,
		"sort": map[string]interface{}{
			"date": map[string]interface{}{
				"order": "desc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"match": map[string]interface{}{"patient": patient}},
					{"match": map[string]interface{}{"doctor": doctor}},
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("[ERROR] Error encoding query: %s", err)
		return nil, err
	}
	res, err := search(es, buf)
	if err != nil {
		log.Fatalf("[ERROR] Error getting response: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("[ERROR] Error parsing the response body: %s", err)
		return nil, err
	}

	// return the history of messages only if they could be found on elasticsearch
	if res.StatusCode != 404 {
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

			// convert it to time.Time because of msg.When field type
			msgDate, _ := time.Parse(hit.(map[string]interface{})["_source"].(map[string]interface{})["date"].(string), "2006-01-02T15:04:05Z")
			aux := msg{
				Message: hit.(map[string]interface{})["_source"].(map[string]interface{})["msg"].(string),
				When:    msgDate,
				Name:    hit.(map[string]interface{})["_source"].(map[string]interface{})["sentBy"].(string),
				Doctor:  hit.(map[string]interface{})["_source"].(map[string]interface{})["doctor"].(string),
				Patient: hit.(map[string]interface{})["_source"].(map[string]interface{})["patient"].(string),
			}

			// store the elasticsearch results in a slice because the dates are in the wrong order
			esMsgs = append(esMsgs, aux)
		}
		return esMsgs, nil
	} else {
		return nil, nil
	}
}
