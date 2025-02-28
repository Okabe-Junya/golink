package mocks

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockFirestoreClient is a mock implementation of the Firestore client
type MockFirestoreClient struct {
	docs map[string]map[string]interface{}
}

// NewMockFirestoreClient creates a new mock Firestore client
func NewMockFirestoreClient() *MockFirestoreClient {
	return &MockFirestoreClient{
		docs: make(map[string]map[string]interface{}),
	}
}

// Collection returns a mock collection reference
func (m *MockFirestoreClient) Collection(path string) *firestore.CollectionRef {
	return &firestore.CollectionRef{}
}

// MockDocumentRef is a mock implementation of the Firestore DocumentRef
type MockDocumentRef struct {
	client *MockFirestoreClient
	path   string
	id     string
}

// MockCollectionRef is a mock implementation of the Firestore CollectionRef
type MockCollectionRef struct {
	client     *MockFirestoreClient
	collection string
}

// Doc returns a mock document reference
func (m *MockCollectionRef) Doc(id string) *MockDocumentRef {
	return &MockDocumentRef{
		client: m.client,
		path:   m.collection,
		id:     id,
	}
}

// Set sets a document in the mock Firestore
func (m *MockDocumentRef) Set(ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	// Convert data to map
	dataMap := make(map[string]interface{})
	// In a real implementation, you would convert the data to a map
	// For simplicity, we'll just store the data as is
	m.client.docs[m.path+"/"+m.id] = dataMap
	return &firestore.WriteResult{}, nil
}

// Get gets a document from the mock Firestore
func (m *MockDocumentRef) Get(ctx context.Context) (*firestore.DocumentSnapshot, error) {
	_, ok := m.client.docs[m.path+"/"+m.id]
	if !ok {
		return nil, status.Error(codes.NotFound, "document not found")
	}
	// Create a mock document snapshot
	// Since we can't directly set the Data field, we'll need to use a different approach
	// In a real test, you would use a more sophisticated mock
	return &firestore.DocumentSnapshot{
		Ref: &firestore.DocumentRef{},
		// We can't set Data directly, so in real tests we'd need to implement DataTo method
	}, nil
}

// Delete deletes a document from the mock Firestore
func (m *MockDocumentRef) Delete(ctx context.Context, opts ...firestore.Precondition) (*firestore.WriteResult, error) {
	delete(m.client.docs, m.path+"/"+m.id)
	return &firestore.WriteResult{}, nil
}

// Documents returns a mock document iterator
func (m *MockCollectionRef) Documents(ctx context.Context) *firestore.DocumentIterator {
	return &firestore.DocumentIterator{}
}

// Where returns a mock query
func (m *MockCollectionRef) Where(path, op string, value interface{}) *firestore.Query {
	return &firestore.Query{}
}

// MockDocumentIterator is a mock implementation of the Firestore DocumentIterator
type MockDocumentIterator struct {
	docs  []*firestore.DocumentSnapshot
	index int
}

// Next returns the next document in the iterator
func (m *MockDocumentIterator) Next() (*firestore.DocumentSnapshot, error) {
	if m.index >= len(m.docs) {
		return nil, iterator.Done
	}
	doc := m.docs[m.index]
	m.index++
	return doc, nil
}
