package grpc

import (
	"context"
	"log"
	"time"

	grpcpkg "google.golang.org/grpc"
)

func LoggingInterceptor() grpcpkg.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpcpkg.UnaryServerInfo,
		handler grpcpkg.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		log.Printf(
			"[payment-service] method=%s duration=%s error=%v",
			info.FullMethod,
			time.Since(start),
			err,
		)

		return resp, err
	}
}