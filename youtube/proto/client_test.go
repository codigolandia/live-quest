package proto

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"testing"
)

var (
	chatId = flag.String("chatid", "", "")
)

func TestClient(t *testing.T) {
	if chatId == nil || *chatId == "" {
		t.Log("Skipping stream chat API client test.")
		t.Skip()
	}
	ctx := context.Background()
	c, err := NewClient(ctx, nil)
	if err != nil {
		t.Fatalf("Error initializing GRPC Client: %v", err)
	}
	nextPageToken := ""
	for {
		req := LiveChatMessageListRequest{
			Part:       []string{"snippet", "authorDetails"},
			LiveChatId: chatId,
			PageToken:  &nextPageToken,
		}
		streamList, err := c.YT.StreamList(ctx, &req)
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
