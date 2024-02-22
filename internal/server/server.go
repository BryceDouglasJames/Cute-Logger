package server

import (
	"errors"
	"fmt"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
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
			fmt.Println("YOO")
			return nil, err
		}
	}

	return srv, nil
}
