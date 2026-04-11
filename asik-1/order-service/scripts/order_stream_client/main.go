package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	orderv1 "github.com/usenbai-nur/ADP2_asik2_generated/order/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./scripts/order_stream_client <order-id>")
	}

	orderID := os.Args[1]

	conn, err := grpc.NewClient(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := orderv1.NewOrderServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stream, err := client.SubscribeToOrderUpdates(ctx, &orderv1.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		log.Fatalf("subscribe error: %v", err)
	}

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed")
			return
		}
		if err != nil {
			log.Fatalf("stream recv error: %v", err)
		}

		log.Printf(
			"order update: order_id=%s status=%s updated_at=%s",
			update.GetOrderId(),
			update.GetStatus().String(),
			update.GetUpdatedAt().AsTime().Format(time.RFC3339),
		)
	}
}