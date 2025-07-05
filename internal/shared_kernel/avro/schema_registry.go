package avro

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// define a schema registry interface
type SchemaRegistry interface {
	GetLatestSchema(subject string) (*SchemaInfo, error)
	GetSchemaByID(schemaID int) (*SchemaInfo, error)
	GetSubjectVersions(subject string) ([]int, error)
	ClearCache()
	GetCachedSchemaCount() int
}

var _ SchemaRegistry = (*SchemaRegistryClient)(nil)

// SchemaRegistryClient provides methods to fetch schemas from the schema registry
type SchemaRegistryClient struct {
	baseURL    string
	httpClient *http.Client
	cache      map[string]*SchemaInfo
	mutex      sync.RWMutex
}

// SchemaInfo represents schema information from the registry
type SchemaInfo struct {
	Subject string `json:"subject"`
	Version int    `json:"version"`
	ID      int    `json:"id"`
	Schema  string `json:"schema"`
}

// NewSchemaRegistryClient creates a new schema registry client
func NewSchemaRegistryClient(baseURL string) *SchemaRegistryClient {
	return &SchemaRegistryClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		cache:      make(map[string]*SchemaInfo),
	}
}

// GetLatestSchema fetches the latest schema for a given subject
func (c *SchemaRegistryClient) GetLatestSchema(subject string) (*SchemaInfo, error) {
	// Check cache first
	c.mutex.RLock()
	if cached, exists := c.cache[subject]; exists {
		c.mutex.RUnlock()
		return cached, nil
	}
	c.mutex.RUnlock()

	// Fetch from registry
	url := fmt.Sprintf("%s/subjects/%s/versions/latest", c.baseURL, subject)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema for subject %s: %w", subject, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch schema for subject %s: status %d, body: %s", subject, resp.StatusCode, string(body))
	}

	var schemaInfo SchemaInfo
	if err := json.NewDecoder(resp.Body).Decode(&schemaInfo); err != nil {
		return nil, fmt.Errorf("failed to decode schema response for subject %s: %w", subject, err)
	}

	// Cache the result
	c.mutex.Lock()
	c.cache[subject] = &schemaInfo
	c.mutex.Unlock()

	return &schemaInfo, nil
}

// GetSchemaByID fetches a schema by its ID
func (c *SchemaRegistryClient) GetSchemaByID(schemaID int) (*SchemaInfo, error) {
	url := fmt.Sprintf("%s/schemas/ids/%d", c.baseURL, schemaID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema by ID %d: %w", schemaID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch schema by ID %d: status %d, body: %s", schemaID, resp.StatusCode, string(body))
	}

	var schemaInfo SchemaInfo
	if err := json.NewDecoder(resp.Body).Decode(&schemaInfo); err != nil {
		return nil, fmt.Errorf("failed to decode schema response for ID %d: %w", schemaID, err)
	}

	return &schemaInfo, nil
}

// GetSubjectVersions fetches all versions for a subject
func (c *SchemaRegistryClient) GetSubjectVersions(subject string) ([]int, error) {
	url := fmt.Sprintf("%s/subjects/%s/versions", c.baseURL, subject)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions for subject %s: %w", subject, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch versions for subject %s: status %d, body: %s", subject, resp.StatusCode, string(body))
	}

	var versions []int
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode versions response for subject %s: %w", subject, err)
	}

	return versions, nil
}

// ClearCache clears the schema cache
func (c *SchemaRegistryClient) ClearCache() {
	c.mutex.Lock()
	c.cache = make(map[string]*SchemaInfo)
	c.mutex.Unlock()
}

// GetCachedSchemaCount returns the number of cached schemas
func (c *SchemaRegistryClient) GetCachedSchemaCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}
