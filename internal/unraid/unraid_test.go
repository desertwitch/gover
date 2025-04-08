package unraid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewHandler_Success tests the factory function.
func TestNewHandler_Success(t *testing.T) {
	mockFS := newMockFsProvider(t)
	mockConfig := newMockConfigProvider(t)
	mockOS := newMockOsProvider(t)

	handler := NewHandler(mockFS, mockConfig, mockOS)

	assert.Equal(t, mockFS, handler.fsHandler)
	assert.Equal(t, mockConfig, handler.configHandler)
	assert.Equal(t, mockOS, handler.osHandler)
}
