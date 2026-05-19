// tasks.md: T063 | spec.md: Milvus client connection
package milvus

import (
	"context"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// Client wraps the Milvus SDK client
type Client struct {
	client.Client
}

// NewClient creates a new Milvus client and verifies the connection
func NewClient(cfg interface{}) (*Client, error) {
	// For now, we'll use a simple configuration approach
	// In production, this should accept the actual MilvusConfig struct
	host := "localhost"
	port := 19530

	// Create Milvus client
	milvusClient, err := client.NewGrpcClient(context.Background(), fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("failed to create Milvus client: %w", err)
	}

	// Verify connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := milvusClient.Ping(ctx); err != nil {
		milvusClient.Close()
		return nil, fmt.Errorf("failed to ping Milvus: %w", err)
	}

	return &Client{Client: milvusClient}, nil
}

// HasCollection checks if a collection exists
func (c *Client) HasCollection(ctx context.Context, collectionName string) (bool, error) {
	return c.Client.HasCollection(ctx, collectionName)
}

// CreateCollection creates a new collection with the specified schema
func (c *Client) CreateCollection(ctx context.Context, collectionName string, schema *entity.Schema, shardNum int32) error {
	return c.Client.CreateCollection(ctx, schema, shardNum)
}

// GetCollection retrieves a collection
func (c *Client) GetCollection(ctx context.Context, collectionName string) (string, error) {
	return c.Client.GetCollection(ctx, collectionName)
}

// DropCollection drops a collection
func (c *Client) DropCollection(ctx context.Context, collectionName string) error {
	return c.Client.DropCollection(ctx, collectionName)
}

// Insert inserts data into a collection
func (c *Client) Insert(ctx context.Context, collectionName string, partitionName string, columns ...entity.Column) error {
	return c.Client.Insert(ctx, collectionName, partitionName, columns...)
}

// Flush flushes data to disk
func (c *Client) Flush(ctx context.Context, collectionNames ...string) error {
	return c.Client.Flush(ctx, collectionNames...)
}

// CreateIndex creates an index on a field
func (c *Client) CreateIndex(ctx context.Context, collectionName string, fieldName string, idx entity.Index, async bool) error {
	return c.Client.CreateIndex(ctx, collectionName, fieldName, idx, async)
}

// DropIndex drops an index
func (c *Client) DropIndex(ctx context.Context, collectionName string, fieldName string) error {
	return c.Client.DropIndex(ctx, collectionName, fieldName)
}

// GetIndexState gets the state of an index
func (c *Client) GetIndexState(ctx context.Context, collectionName string, fieldName string) (entity.IndexState, error) {
	return c.Client.GetIndexState(ctx, collectionName, fieldName)
}

// LoadCollection loads a collection into memory
func (c *Client) LoadCollection(ctx context.Context, collectionName string, async bool) error {
	return c.Client.LoadCollection(ctx, collectionName, async)
}

// Search performs a vector similarity search
func (c *Client) Search(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, vectors []entity.Vector, vectorField string, metricType entity.MetricType, topK int, sp entity.SearchParam) ([]entity.SearchResult, error) {
	return c.Client.Search(ctx, collectionName, partitions, expr, outputFields, vectors, vectorField, metricType, topK, sp)
}

// Query performs a query with expression
func (c *Client) Query(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, limit int64) ([]entity.Column, error) {
	return c.Client.Query(ctx, collectionName, partitions, expr, outputFields, limit)
}

// Delete deletes entities matching the expression
func (c *Client) Delete(ctx context.Context, collectionName string, expr string) error {
	return c.Client.Delete(ctx, collectionName, expr)
}
