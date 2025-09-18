package infrastructure

import (
	"blog-api/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/olivere/elastic/v7"
)

// SearchClient wraps Elasticsearch operations
type SearchClient struct {
	client *elastic.Client
    ctx    context.Context
}

// NewSearchClient creates a new Elasticsearch search client
func NewSearchClient(cfg *config.Config) (*SearchClient, error) {
	ctx := context.Background()

	// Create Elasticsearch client
	es, err := elastic.NewClient(
		elastic.SetURL(cfg.Elasticsearch.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(100, 200))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %v", err)
	}

	// Test connection
	info, code, err := es.Ping(cfg.Elasticsearch.URL).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %v", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("Elasticsearch ping returned code %d", code)
	}

	log.Printf("✅ Elasticsearch connected successfully (version: %s)", info.Version.Number)

	return &SearchClient{
		client: es,
        ctx:   ctx,
	}, nil
}

// GetClient returns the Elasticsearch client
func (sc *SearchClient) GetClient() *elastic.Client {
	return sc.client
}

// GetContext returns the context
func (sc *SearchClient) GetContext() context.Context {
	return sc.ctx
}

// Close closes the Elasticsearch connection
func (sc *SearchClient) Close() error {
	// Elasticsearch client doesn't have a Close method in v7
	if sc.client != nil {
		log.Printf("✅ Elasticsearch connection released")
	}
	return nil
}

// Health checks Elasticsearch health
func (sc *SearchClient) Health() bool {
	if sc.client == nil {
		return false
	}
	_, code, err := sc.client.Ping("http://localhost:9200").Do(sc.ctx)
	return err == nil && code == 200
}

// Stats returns Elasticsearch statistics
func (sc *SearchClient) Stats() map[string]interface{} {
	if sc.client == nil {
		return map[string]interface{}{"status": "disconnected"}
	}

	// For now, return basic status. Can be extended with cluster stats
	return map[string]interface{}{
		"status": "connected",
		"url":    "elasticsearch_cluster",
	}
}

// InitializePostsIndex creates the posts index if it doesn't exist
func (sc *SearchClient) InitializePostsIndex() error {
	indexName := "posts"

	exists, err := sc.client.IndexExists(indexName).Do(sc.ctx)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %v", err)
	}

	if !exists {
		mapping := `{
			"mappings": {
				"properties": {
					"id": {
						"type": "integer"
					},
					"title": {
						"type": "text",
						"analyzer": "standard",
						"fields": {
							"keyword": {
								"type": "keyword",
								"ignore_above": 256
							}
						}
					},
					"content": {
						"type": "text",
						"analyzer": "standard"
					},
					"tags": {
						"type": "keyword"
					},
					"created_at": {
						"type": "date"
					},
					"updated_at": {
						"type": "date"
					}
				}
			},
			"settings": {
				"number_of_shards": 1,
				"number_of_replicas": 0,
				"analysis": {
					"analyzer": {
						"custom_text_analyzer": {
							"type": "standard",
							"stopwords": "_english_"
						}
					}
				}
			}
		}`

		createIndex, err := sc.client.CreateIndex(indexName).Body(mapping).Do(sc.ctx)
		if err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}

		if !createIndex.Acknowledged {
			return fmt.Errorf("index creation not acknowledged")
		}

		log.Printf("✅ Elasticsearch index '%s' created successfully", indexName)
	} else {
		log.Printf("✅ Elasticsearch index '%s' already exists", indexName)
	}

	return nil
}
