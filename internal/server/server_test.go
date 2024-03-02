package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

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
		//"produce/consume a message to/from the log succeeds":           testProduceConsume,
		//"raw gRPC server produce and consume":                          testRawGrpcServerProduceAndConsume,
		//"testing gRPC produce stream with a mock server":               testProduceStreamWithMockServer,
		//"raw gRPC server streaming produce and consume":                testRawGrpcServerStreamProduceAndConsume,
		"raw gRPC server streaming stress test on produce and consume": testRawGrpcServerStreamProduceAndConsumeStressTest,
	} {
		t.Run(scenario, func(t *testing.T) {
			t.Log("YOOOO")
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

func testRawGrpcServerStreamProduceAndConsume(t *testing.T, client api.LogClient, ctx context.Context) {
	// Define a slice of records to send through the ProduceStream
	records := []*api.Record{
		{Value: []byte("first message")},
		{Value: []byte("second message")},
	}

	// Open a stream to produce records
	produceStream, err := client.ProduceStream(ctx)
	require.NoError(t, err)

	// Send records and receive responses to verify offsets
	for _, record := range records {
		err := produceStream.Send(&api.ProduceRequest{Record: record})
		require.NoError(t, err)
		res, err := produceStream.Recv()
		require.NoError(t, err)
		record.Offset = res.Offset // Update record with the offset received
	}
	err = produceStream.CloseSend()
	require.NoError(t, err)

	// Open a stream to consume records starting from the first offset
	consumeStream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: records[0].Offset})
	require.NoError(t, err)

	// Receive and verify records from the consume stream
	for i, want := range records {
		res, err := consumeStream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		// Verify that the consumed record matches the expected record
		require.Equal(t, want.Value, res.Record.Value, fmt.Sprintf("Record %d value mismatch", i))
		require.Equal(t, want.Offset, res.Record.Offset, fmt.Sprintf("Record %d offset mismatch", i))
	}
}

/* I <3 concurrent programming
* I have spent way too much time on this
* Having issues reading records back while retaining offset positions.*/
func testRawGrpcServerStreamProduceAndConsumeStressTest(t *testing.T, client api.LogClient, ctx context.Context) {
	recordCount := 500 // Number of records for the stress test
	workers := 10      // Number of concurrent workers
	var wg sync.WaitGroup
	recordsPerWorker := recordCount / workers

	// Start time for consumer setup
	startTime := time.Now()

	// Set up consumer
	consumeStream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{})
	require.NoError(t, err)

	consumeCount := 0
	go func() {
		for {
			_, err := consumeStream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
	}()

	// Start producing records after consumer setup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			produceStream, err := client.ProduceStream(ctx)
			require.NoError(t, err)

			for j := 0; j < recordsPerWorker; j++ {
				record := &api.Record{Value: []byte(fmt.Sprintf("message %d from worker %d", j, workerID))}
				err := produceStream.Send(&api.ProduceRequest{Record: record})
				require.NoError(t, err)

				res, err := produceStream.Recv()
				require.NoError(t, err)
				require.GreaterOrEqual(t, res.Offset, uint64(0), "Expected valid offset")
			}
			err = produceStream.CloseSend()
			require.NoError(t, err)
		}(i)
	}

	wg.Wait() // Wait for all production to complete

	// Measure time taken to produce records
	produceDuration := time.Since(startTime)
	fmt.Printf("Produced %d records with %d workers in %v\n", recordCount, workers, produceDuration)

	// Ensure all records were consumed
	require.Equal(t, recordCount, consumeCount, "Expected to consume the same number of records as produced")

	// Measure total time taken for the test
	totalDuration := time.Since(startTime)
	fmt.Printf("Stress test completed: produced and consumed %d records in %v\n", recordCount, totalDuration)
}
