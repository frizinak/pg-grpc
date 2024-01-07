.PHONY: all
all: pb/grpc.pb.go

pb/grpc.pb.go: pb/app.proto
	cd pb && \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			app.proto
