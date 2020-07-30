package pb

import (
	"log"
	"net"

	"github.com/alash3al/redix/db"
	grpc "google.golang.org/grpc"
)

func ListenAndServe(addr string, store *db.DB) error {
	lis, err := net.Listen("tcp", ":3035")
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}

	server := grpc.NewServer()

	RegisterRedixServiceServer(server, &RedixService{db: store})

	return server.Serve(lis)
}
