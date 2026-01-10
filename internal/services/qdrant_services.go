package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"dory-backend/internal/config"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/qdrant/go-client/qdrant"
)

var QClient *qdrant.Client

func InitQdrant() {
	var err error
	host := config.AppConfig.QdrantHost
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")

	QClient, err = qdrant.NewClient(&qdrant.Config{
		Host:   host,
		Port:   6334,
		APIKey: config.AppConfig.QdrantKey,
		UseTLS: true,
	})

	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	log.Println("Qdrant Client connected successfully to:", host)

	ctx := context.Background()
	collectionName := "user_text_embeddings"

	err = QClient.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1024,
			Distance: qdrant.Distance_Cosine,
		}),
	})

	if err != nil {
		log.Println("Collection might already exist or: ", err)
	} else {
		log.Println("Qdrant Collection 'user_text_embeddings' initialized.")
	}

	// Create index via REST API (more reliable than gRPC enums)
	createFieldIndexViaREST(host, collectionName, "user_id")
}

// createFieldIndexViaREST uses the REST API to create a keyword index on user_id
func createFieldIndexViaREST(host string, collectionName string, fieldName string) {
	url := fmt.Sprintf("https://%s:6333/collections/%s/index", host, collectionName)

	payload := map[string]interface{}{
		"field_name": fieldName,
		"field_type": "keyword",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.AppConfig.QdrantKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to create index via REST:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		log.Printf("Successfully created keyword index for '%s' field", fieldName)
	} else {
		log.Printf("Index creation returned status %d (might already exist)", resp.StatusCode)
	}
}

func StoreChunksInQdrant(userID string, docID string, chunks []string) error {
	var points []*qdrant.PointStruct

	for i, text := range chunks {
		vector, err := EmbedText(text)
		if err != nil {
			log.Printf("Embedding failed for chunk %d of doc %s: %v", i, docID, err)
			continue
		}

		// Generate unique ID for each chunk using docID and chunk index
		// This ensures no collisions across different documents
		uniqueID := fmt.Sprintf("%s_%d", docID, i)
		hash := md5.Sum([]byte(uniqueID))
		// Use first 8 bytes of hash as uint64 ID
		idNum := uint64(hash[0]) | uint64(hash[1])<<8 | uint64(hash[2])<<16 | uint64(hash[3])<<24 |
			uint64(hash[4])<<32 | uint64(hash[5])<<40 | uint64(hash[6])<<48 | uint64(hash[7])<<56

		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(idNum),
			Vectors: qdrant.NewVectors(vector...),
			Payload: qdrant.NewValueMap(map[string]any{
				"user_id":     userID,
				"document_id": docID,
				"chunk_index": int64(i),
				"content":     text,
			}),
		}
		points = append(points, point)
	}

	if len(points) == 0 {
		return fmt.Errorf("no chunks were successfully embedded for document %s", docID)
	}

	_, err := QClient.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: "user_text_embeddings",
		Points:         points,
		Wait:           boolPtr(true),
	})

	if err != nil {
		return fmt.Errorf("failed to upsert points to Qdrant for doc %s: %v", docID, err)
	}

	log.Printf("Successfully stored %d chunks in Qdrant for document %s", len(points), docID)
	return nil
}
func boolPtr(b bool) *bool {
	return &b
}

func SearchSimilarChunks(userID string, queryText string) ([]string, error) {
	queryVector, err := EmbedText(queryText)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	searchResponse, err := QClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "user_text_embeddings",
		Query:          qdrant.NewQuery(queryVector...),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatch("user_id", userID),
			},
		},
		Limit:       uint64Ptr(5),
		WithPayload: qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, err
	}

	var results []string
	for _, hit := range searchResponse {
		if content, ok := hit.Payload["content"]; ok {
			results = append(results, content.GetStringValue())
		}
	}

	return results, nil
}

func uint64Ptr(i uint64) *uint64 { return &i }
