package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch"
	"log"
	"time"
)

func logToElasticsearch(logMessage string) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elasticsearch:9200",
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	logEntry := map[string]interface{}{
		"message":   logMessage,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(logEntry)
	if err != nil {
		log.Fatalf("Error marshaling log entry: %s", err)
	}

	res, err := es.Index(
		"logs",
		bytes.NewReader(body),
		es.Index.WithContext(context.Background()),
		es.Index.WithDocumentType("log"),
		es.Index.WithRefresh("true"),
	)

	if err != nil {
		log.Fatalf("Error indexing log entry: %s", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error indexing log entry: %s", res.String())
	} else {
		log.Printf("Successfully indexed log entry: %s", res.String())
	}
}
