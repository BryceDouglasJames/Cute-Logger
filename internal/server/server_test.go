package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGRPCServer(t *testing.T) {
	// Create server with options
	server, err := NewGRPCServer()
	require.NoError(t, err)
	require.NotNil(t, server)
}
