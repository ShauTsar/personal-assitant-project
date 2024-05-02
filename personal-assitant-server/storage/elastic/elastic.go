package elastic

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"net/http"
	"os"
	"time"
)

func LogToElasticsearch(logMessage string) {
	esHost := os.Getenv("ELASTICSEARCH_HOST")
	if esHost == "" {
		esHost = "https://10.150.0.53:9200" // Значение по умолчанию для имени Kubernetes Service
	}

	apiKey := os.Getenv("ELASTICSEARCH_API_KEY")
	if apiKey == "" {
		apiKey = "ajdKMXZJNEJzbk5NN21MeVlwVUI6dEZuM0hrS0VUdktIRWtlN1B4Nkt0dw=="
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			esHost,
		},
		APIKey: apiKey,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Пропустить проверку сертификата (не рекомендуется в продакшене)
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
