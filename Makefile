.PHONY: test compile

test:
	go test -race ./... -coverprofile=coverage.txt

compile:
	protoc -I api/ api/record.proto --go_out=api --go-grpc_out=api --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative