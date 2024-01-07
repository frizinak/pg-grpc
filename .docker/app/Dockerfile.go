FROM ubuntu:24.04

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -q -y golang ca-certificates protobuf-compiler

COPY go.mod go.sum ./

ENV GOPATH=/go-cache
ENV GOMODCACHE=/go-cache/mod
ENV GOCACHE=/go-cache/cache
ENV PATH="$PATH:$GOPATH/bin"

RUN go mod download

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

COPY . .

RUN cd pb && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		app.proto

ENTRYPOINT ["true"]
