package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/frizinak/pg-grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ex(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "-- CLIENT:", err)
	os.Exit(1)
}

func main() {
	mtries := 10
	var c pb.AppClient
	for {
		conn, err := grpc.Dial("docker.server:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			defer conn.Close()
			c = pb.NewAppClient(conn)
			break
		}

		time.Sleep(time.Millisecond * 200)
		if mtries--; mtries == 0 {
			ex(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Pages(ctx, &pb.PagesRequest{})
	ex(err)
	for {
		page, err := r.Recv()
		if page == nil && err == io.EOF {
			break
		}
		ex(err)
		fmt.Printf("-- CLIENT: %+v\n", page)
	}
}
