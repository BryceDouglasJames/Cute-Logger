package server

import (
	"context"
	"io"
	"net"
	"os"
	"testing"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	log "github.com/BryceDouglasJames/Cute-Logger/internal/logger"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, client api.LogClient, ctx context.Context){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"raw gRPC server produce and consume":                testRawGrpcServerProduceAndConsume,
		"testing gRPC produce stream with a mock server":     testProduceStreamWithMockServer,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, teardown := setupTest(t, nil)
			defer teardown()
			ctx := context.Background()
			fn(t, client, ctx)
		})
	}
}

// setupTest prepares the environment for testing the gRPC Log service.
func setupTest(t *testing.T, fn func(*Config)) (client api.LogClient, teardown func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	ctx := context.Background()

	server, _, err := initializeServer(ctx, lis, fn)
	require.NoError(t, err)

	cc, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(
		func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client = api.NewLogClient(cc)

	teardown = func() {
		cc.Close()
		lis.Close()
		server.GracefulStop()
	}

	return client, teardown
}

// initializeServer sets up and starts the gRPC server.
func initializeServer(ctx context.Context, lis *bufconn.Listener, fn func(*Config)) (server *grpc.Server, cfg *Config, err error) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test")
	if err != nil {
		return nil, nil, err
	}

	// Use CommitLog interface with our logger
	clog, err := log.NewLog(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, err
	}

	server = grpc.NewServer()
	grpcServer, err := NewGRPCServer(WithCommitLog(clog))
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, err
	}

	api.RegisterLogServer(server, grpcServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			panic("Failed to serve: " + err.Error())
		}
	}()

	return server, cfg, nil
}

func testRawGrpcServerProduceAndConsume(t *testing.T, _ api.LogClient, ctx context.Context) {
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
	produceResp, err := server.Produce(ctx, &api.ProduceRequest{Record: record})
	require.NoError(t, err)
	require.NotNil(t, produceResp)
	require.Equal(t, uint64(0), produceResp.Offset)

	// Test Consume with the offset received from Produce
	consumeResp, err := server.Consume(ctx, &api.ConsumeRequest{Offset: produceResp.Offset})
	require.NoError(t, err)
	require.NotNil(t, consumeResp)
	require.Equal(t, record.Value, consumeResp.Record.Value)
}

func testProduceConsume(t *testing.T, client api.LogClient, ctx context.Context) {
	// Test Produce
	record := &api.Record{Value: []byte("test record")}
	produceResp, err := client.Produce(ctx, &api.ProduceRequest{Record: record})
	require.NoError(t, err)
	require.NotNil(t, produceResp)
	require.Equal(t, uint64(0), produceResp.Offset)

	// Test Consume with the offset received from Produce
	consumeResp, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produceResp.Offset})
	require.NoError(t, err)
	require.NotNil(t, consumeResp)
	require.Equal(t, record.Value, consumeResp.Record.Value)
}

func testProduceStreamWithMockServer(t *testing.T, _ api.LogClient, _ context.Context) {
	// This uses gomock to simulate incoming stream requests and validate the
	// behavior of the server in handling streaming data production.
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_stream")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new Log instance with the temporary directory
	clog, err := log.NewLog(tempDir)
	require.NoError(t, err)

	// Initialize a new mock controller with the current testing context
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a new mock instance of the ProduceStreamServer to simulate client requests
	mockStream := NewMockLog_ProduceStreamServer(ctrl)

	// Create server with options
	// Create a new gRPC server with the default configuration.
	gsrv := grpc.NewServer()
	defer gsrv.Stop()

	// Create server and attach our logger interface
	server, err := NewGRPCServer(WithCommitLog(clog))
	require.NoError(t, err)
	require.NotNil(t, server)
	api.RegisterLogServer(gsrv, server)

	// Define a request with a sample record to be sent to the server
	req := &api.ProduceRequest{
		Record: &api.Record{Value: []byte("test record swag")},
	}

	// Define the expected response from the server after processing the request
	res := &api.ProduceResponse{Offset: 0}

	// Set up the expected sequence of interactions between the test and the mock stream.
	// This includes receiving a request, getting the context, sending a response, and simulating the end of the stream.
	gomock.InOrder(
		mockStream.EXPECT().Recv().Return(req, nil),
		mockStream.EXPECT().Context().Return(context.Background()),
		mockStream.EXPECT().Send(res).Return(nil),
		mockStream.EXPECT().Recv().Return(nil, io.EOF),
	)

	// Call the ProduceStream method with the mocked stream and check for errors
	err = server.ProduceStream(mockStream)
	if err != nil {
		t.Fatalf("ProduceStream failed: %v", err)
	}
}
