package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stackyrd/config"
	"stackyrd/pkg/logger"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/viper"
)

// QdrantManager manages Qdrant vector database interactions
type QdrantManager struct {
	Client  *retryablehttp.Client
	BaseURL string
	APIKey  string
	Pool    *WorkerPool // Async worker pool
	logger  *logger.Logger
}

// QdrantPoint represents a vector point in Qdrant
type QdrantPoint struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// QdrantSearchRequest represents a search query
type QdrantSearchRequest struct {
	Vector      []float32              `json:"vector"`
	Top         int                    `json:"top"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
	WithPayload bool                   `json:"with_payload,omitempty"`
	WithVector  bool                   `json:"with_vector,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// QdrantSearchResult represents a search result
type QdrantSearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	Vector  []float32              `json:"vector,omitempty"`
}

// QdrantCollectionInfo represents collection information
type QdrantCollectionInfo struct {
	Name        string `json:"name"`
	VectorSize  int    `json:"vector_size"`
	Distance    string `json:"distance"`
	PointsCount int    `json:"points_count,omitempty"`
	Status      string `json:"status,omitempty"`
}

// qdrantLoggerAdapter adapts our custom logger
type qdrantLoggerAdapter struct {
	logger *logger.Logger
}

func (a *qdrantLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	a.logger.Error(msg, nil, keysAndValues...)
}

func (a *qdrantLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.logger.Info(msg, keysAndValues...)
}

func (a *qdrantLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	a.logger.Debug(msg, keysAndValues...)
}

func (a *qdrantLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	a.logger.Warn(msg, keysAndValues...)
}

// Name returns the display name of the component
func (q *QdrantManager) Name() string {
	return "Qdrant Vector Database"
}

// NewQdrantManager creates a new Qdrant manager
func NewQdrantManager(logger *logger.Logger) (*QdrantManager, error) {
	enabled := viper.GetBool("qdrant.enabled")
	if !enabled {
		return nil, nil
	}

	host := viper.GetString("qdrant.host")
	port := viper.GetInt("qdrant.port")
	apiKey := viper.GetString("qdrant.api_key")

	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	logger.Info("Initializing Qdrant manager", "url", baseURL)

	// Create HTTP client with retry logic
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = time.Second
	client.RetryWaitMax = 10 * time.Second
	client.HTTPClient.Timeout = 60 * time.Second

	// Set custom logger
	client.Logger = &qdrantLoggerAdapter{logger: logger}

	manager := &QdrantManager{
		Client:  client,
		BaseURL: baseURL,
		APIKey:  apiKey,
		logger:  logger,
	}

	// Test connection
	if err := manager.testConnection(); err != nil {
		logger.Error("Qdrant connection test failed", err)
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	logger.Info("Qdrant connection test successful")

	// Initialize worker pool for async operations
	pool := NewWorkerPool(6)
	pool.Start()

	manager.Pool = pool
	logger.Info("Qdrant manager initialized with worker pool")

	return manager, nil
}

// testConnection tests the connection to Qdrant
func (q *QdrantManager) testConnection() error {
	req, err := retryablehttp.NewRequest("GET", q.BaseURL+"/", nil)
	if err != nil {
		return err
	}

	if q.APIKey != "" {
		req.Header.Set("api-key", q.APIKey)
	}

	resp, err := q.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		q.logger.Error("Qdrant health check failed", nil, "status", resp.StatusCode)
		return fmt.Errorf("Qdrant health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// UpsertPoints inserts or updates multiple points in a collection
func (q *QdrantManager) UpsertPoints(ctx context.Context, collection string, points []QdrantPoint) error {
	q.logger.Debug("Upserting points", "collection", collection, "count", len(points))

	payload := map[string]interface{}{
		"points": points,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal points: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/collections/%s/points", q.BaseURL, collection), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if q.APIKey != "" {
		req.Header.Set("api-key", q.APIKey)
	}

	resp, err := q.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert points failed: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// Search performs a vector search
func (q *QdrantManager) Search(ctx context.Context, collection string, request QdrantSearchRequest) ([]QdrantSearchResult, error) {
	q.logger.Debug("Searching vectors", "collection", collection, "top", request.Top)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/collections/%s/points/search", q.BaseURL, collection), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if q.APIKey != "" {
		req.Header.Set("api-key", q.APIKey)
	}

	resp, err := q.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Result []QdrantSearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return result.Result, nil
}

// GetStatus returns the current status of the Qdrant manager
func (q *QdrantManager) GetStatus() map[string]interface{} {
	stats := make(map[string]interface{})
	if q == nil {
		stats["connected"] = false
		return stats
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", q.BaseURL+"/", nil)
	if err != nil {
		stats["connected"] = false
		stats["error"] = err.Error()
		return stats
	}

	if q.APIKey != "" {
		req.Header.Set("api-key", q.APIKey)
	}

	resp, err := q.Client.Do(req)
	if err != nil {
		stats["connected"] = false
		stats["error"] = err.Error()
		return stats
	}
	defer resp.Body.Close()

	stats["connected"] = resp.StatusCode == http.StatusOK
	stats["base_url"] = q.BaseURL

	if q.Pool != nil {
		stats["pool_active"] = true
	}

	return stats
}

// Async Operations

// UpsertPointsAsync asynchronously upserts points
func (q *QdrantManager) UpsertPointsAsync(ctx context.Context, collection string, points []QdrantPoint) *AsyncResult[struct{}] {
	return ExecuteAsync(ctx, func(ctx context.Context) (struct{}, error) {
		err := q.UpsertPoints(ctx, collection, points)
		return struct{}{}, err
	})
}

// SearchAsync asynchronously performs vector search
func (q *QdrantManager) SearchAsync(ctx context.Context, collection string, request QdrantSearchRequest) *AsyncResult[[]QdrantSearchResult] {
	return ExecuteAsync(ctx, func(ctx context.Context) ([]QdrantSearchResult, error) {
		return q.Search(ctx, collection, request)
	})
}

// SubmitAsyncJob submits an async job to the worker pool
func (q *QdrantManager) SubmitAsyncJob(job func()) {
	if q.Pool != nil {
		q.Pool.Submit(job)
	} else {
		go job()
	}
}

// Close closes the Qdrant manager and its worker pool
func (q *QdrantManager) Close() error {
	if q.Pool != nil {
		q.Pool.Close()
	}
	return nil
}

func init() {
	RegisterComponent("qdrant", func(cfg *config.Config, l *logger.Logger) (InfrastructureComponent, error) {
		return NewQdrantManager(l)
	})
}
