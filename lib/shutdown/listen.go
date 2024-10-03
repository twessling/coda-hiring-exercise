package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func ListenStopSignal(ctx context.Context, cancelFunc context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		log.Println("Stop signal received, shutting down service...")
		cancelFunc()
	case <-ctx.Done():
	}
}
