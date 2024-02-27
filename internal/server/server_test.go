package server

import (
	"context"
	"os"
	"testing"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	log "github.com/BryceDouglasJames/Cute-Logger/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestNewGRPCServer(t *testing.T) {
	// Create server with options
	server, err := NewGRPCServer()
	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestGrpcServerProduceAndConsume(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_again")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new Log instance with the temporary directory
	clog, err := log.NewLog(tempDir)
	require.NoError(t, err)

	// Initialize grpcServer with in-memory commit log
	server, err := NewGRPCServer(WithCommitLog(clog))
	require.NoError(t, err)

	// Test Produce
	record := &api.Record{Value: []byte("test record")}
	produceResp, err := server.Produce(context.Background(), &api.ProduceRequest{Record: record})
	require.NoError(t, err)
	require.NotNil(t, produceResp)
	require.Equal(t, uint64(0), produceResp.Offset)

	// Test Consume with the offset received from Produce
	consumeResp, err := server.Consume(context.Background(), &api.ConsumeRequest{Offset: produceResp.Offset})
	require.NoError(t, err)
	require.NotNil(t, consumeResp)
	require.Equal(t, record.Value, consumeResp.Record.Value)
}
