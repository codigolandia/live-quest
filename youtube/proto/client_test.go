package proto

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"testing"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcoauth "google.golang.org/grpc/credentials/oauth"

	"github.com/codigolandia/live-quest/oauth"
)

var (
	serverAddr  = "dns:///youtube.googleapis.com:443"
	systemCreds = credentials.NewTLS(&tls.Config{})
	chatId      = flag.String("chatid", "", "")
)

func TestClient(t *testing.T) {
	if chatId == nil || *chatId == "" {
		t.Log("Skipping stream chat API client test.")
		t.Skip()
	}
	ts, err := oauth.NewTokenSource("Youtube")
	rpcCreds := grpcoauth.TokenSource{TokenSource: ts}
	ctx := context.Background()
	opts := make([]grpc.DialOption, 0)
	opts = append(opts,
		grpc.WithTransportCredentials(systemCreds),
		grpc.WithPerRPCCredentials(rpcCreds),
	)

	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		t.Fatalf("Failed to initialize connection: %v", err)
	}
	nextPageToken := ""
	for {
		client := NewV3DataLiveChatMessageServiceClient(conn)
		req := LiveChatMessageListRequest{
			Part:       []string{"snippet", "authorDetails"},
			LiveChatId: chatId,
			PageToken:  &nextPageToken,
		}

		streamList, err := client.StreamList(ctx, &req)
		if err != nil {
			t.Fatalf("Failed to start streaming chat messages")
		}
		count := 0
		for {
			resp, err := streamList.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error receiving message: %v", err)
			}
			nextPageToken = resp.GetNextPageToken()
			for _, item := range resp.Items {
				author := item.GetAuthorDetails()
				snip := item.GetSnippet()
				count++
				fmt.Printf("[%03d] %v: %v\n",
					count,
					author.GetDisplayName(),
					snip.GetTextMessageDetails().GetMessageText())
			}
		}
	}
}
