package proto

import (
	context "context"
	"crypto/tls"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcoauth "google.golang.org/grpc/credentials/oauth"
)

var (
	systemCreds = credentials.NewTLS(&tls.Config{})
	serverAddr  = "dns:///youtube.googleapis.com:443"
)

type DataClient struct {
	conn *grpc.ClientConn
	YT   V3DataLiveChatMessageServiceClient
}

func NewClient(ctx context.Context, ts oauth2.TokenSource) (c *DataClient, err error) {
	c = new(DataClient)
	rpcCreds := grpcoauth.TokenSource{TokenSource: ts}
	opts := make([]grpc.DialOption, 0)
	opts = append(opts,
		grpc.WithTransportCredentials(systemCreds),
		grpc.WithPerRPCCredentials(rpcCreds),
	)

	c.conn, err = grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize connection: %v", err)
	}
	c.YT = NewV3DataLiveChatMessageServiceClient(c.conn)

	return
}
