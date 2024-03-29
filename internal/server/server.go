package server

import (
	"context"
	"errors"
	"io"
	"log"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CommitLog defines the interface for a commit log system.
// It's designed to abstract the underlying operations of appending to
// and reading from a log, allowing for different implementations that
// could optimize for various use cases (e.g., performance, durability).
type CommitLog interface {
	// Append adds a new record to the log and returns the offset
	// at which the record was stored. This offset can be used to
	// retrieve the record later. An error is returned if the append
	// operation fails.
	Append(*api.Record) (uint64, error)

	// Read retrieves a record from the log at the specified offset.
	// It returns the record if found or an error if the read operation
	// fails, including if the offset does not correspond to an existing
	// record.
	Read(uint64) (*api.Record, error)
}

// Config represents the configuration for the server
type Config struct {
	CommitLog CommitLog
}

// Ensure grpcServer implements the LogServer interface
var _ api.LogServer = (*grpcServer)(nil)

// grpcServer wraps the gRPC server and its configuration
type grpcServer struct {
	api.UnimplementedLogServer
	*Config
}

// Option defines a function signature for configuring the grpcServer
type Option func(*grpcServer) error

// Configures the server to use a specific CommitLog implementation.
// This option allows the server's behavior to be modified based on the provided commit log.
func WithCommitLog(cl CommitLog) Option {
	// return function handle assigning fields
	return func(s *grpcServer) error {
		if cl == nil {
			return errors.New("CommitLog cannot be nil")
		}
		s.Config.CommitLog = cl
		return nil
	}
}

// NewGRPCServer initializes and returns a new grpcServer instance.
// It takes functional options that modify its configuration.
func NewGRPCServer(opts ...Option) (*grpcServer, error) {

	// Initialize the server with default configuration
	srv := &grpcServer{
		Config: &Config{},
	}

	// Apply each Option passed to the function
	for _, opt := range opts {
		err := opt(srv)
		if err != nil {
			return nil, err
		}
	}

	return srv, nil
}

// Produce handles the gRPC call for producing (appending) a record to the commit log
func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	// Validate the incoming request
	if req == nil || req.Record == nil {
		log.Println("Invalid request: request or request record is nil")
		return nil, status.Errorf(codes.InvalidArgument, "request and request record must not be nil")
	}

	// Use the context to support cancellation and deadlines
	select {
	case <-ctx.Done():
		log.Println("Request cancelled or deadline exceeded")
		return nil, ctx.Err()
	default:
		// Continue if the context is not done
	}

	// Append the record contained in the request to the commit log
	offset, err := s.CommitLog.Append(req.Record) // Assuming Append supports context
	if err != nil {
		log.Printf("Error appending to commit log: %v", err)
		return nil, status.Errorf(codes.Internal, "error appending to commit log: %v", err)
	}

	// If the append is successful, construct and return a ProduceResponse with the offset of the appended record
	response := &api.ProduceResponse{Offset: offset}
	log.Printf("Record appended to commit log at offset %d", offset)
	return response, nil
}

func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		// Attempt to receive a message from the stream
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Stream closed by client")
				return nil
			}
			log.Printf("Error receiving from stream: %v\n", err)
			return status.Errorf(codes.Unknown, "Error receiving from stream: %v", err)
		}

		// Log the received request for debugging purposes
		log.Printf("Received request: %v\n", req)

		// Call the Produce method to process the received request
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			log.Printf("Error producing message: %v\n", err)
			return status.Errorf(codes.Internal, "Error producing message: %v", err)
		}

		// Attempt to send the response back to the client
		if err = stream.Send(res); err != nil {
			log.Printf("Error sending to stream: %v\n", err)
			return status.Errorf(codes.Unknown, "Error sending to stream: %v", err)
		}

		// Log the sent response for debugging purposes
		log.Printf("Sent response: %v\n", res)
	}
}

// Consume handles the gRPC call for consuming (reading) a record from the commit log
func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {

	// Read the record from the commit log at the specified offset in the request
	record, err := s.CommitLog.Read(req.Offset)

	// If there's an error reading the record, return the error immediately
	if err != nil {
		return nil, err
	}

	// If the read is successful, return a ConsumeResponse with the read record
	return &api.ConsumeResponse{Record: record}, nil
}

// ConsumeStream streams log entries starting from the requested offset
func (s *grpcServer) ConsumeStream(req *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		// Check if the stream's context is done/cancelled
		case <-stream.Context().Done():

			// Stream is done, so return without error
			return nil

		default:
			// Attempt to consume a log entry at the current offset
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil: // No error, proceed
			default: // Any other error, return it
				return err
			}

			// Send the consumed log entry back to the client
			if err = stream.Send(res); err != nil {
				return err // Error sending to stream, return the error
			}

			req.Offset++ // Increment the offset for the next iteration/request
		}
	}
}
