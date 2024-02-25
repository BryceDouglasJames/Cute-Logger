package server

import (
	"context"
	"errors"
	"testing"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	"github.com/stretchr/testify/require"
)

// **--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**
// Declare in memory example of the commit log for testing purposes
type inMemoryCommitLog struct {
	records []*api.Record
}

func newInMemoryCommitLog() *inMemoryCommitLog {
	return &inMemoryCommitLog{
		records: []*api.Record{},
	}
}

func (log *inMemoryCommitLog) Append(record *api.Record) (uint64, error) {
	offset := uint64(len(log.records))
	log.records = append(log.records, record)
	return offset, nil
}

func (log *inMemoryCommitLog) Read(offset uint64) (*api.Record, error) {
	if offset >= uint64(len(log.records)) {
		return nil, errors.New("offset out of bounds")
	}
	return log.records[offset], nil
}

// **--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**--**

func TestNewGRPCServer(t *testing.T) {
	// Create server with options
	server, err := NewGRPCServer()
	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestGrpcServerProduceAndConsume(t *testing.T) {
	// Setup in-memory commit log
	commitLog := newInMemoryCommitLog()

	// Initialize grpcServer with in-memory commit log
	server, err := NewGRPCServer(WithCommitLog(commitLog))
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

// TODO: Gotta fix this and figure out GoMock before moving on
/*
func TestProduceStream(t *testing.T) {
	// This uses gomock to simulate incoming stream requests and validate the behavior of the server in handling streaming data production.

	// Initialize a new mock controller with the current testing context
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a new mock instance of the ProduceStreamServer to simulate client requests
	mockStream := NewMockLog_ProduceStreamServer(ctrl)

	// Create server with options
	s, err := NewGRPCServer()
	require.NoError(t, err)
	require.NotNil(t, s)

	// Define a request with a sample record to be sent to the server
	req := &api.ProduceRequest{
		Record: &api.Record{Value: []byte("test data")},
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
	// err = s.ProduceStream(mockStream)
	// if err != nil {
	//	t.Fatalf("ProduceStream failed: %v", err)
	// }
}
*/
