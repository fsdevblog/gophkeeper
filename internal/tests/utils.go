package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"resty.dev/v3"
	"time"
)

func mustStartHTTPServer(ctx context.Context, addr string, router *gin.Engine) {
	lis, errLis := net.Listen("tcp", addr)
	if errLis != nil {
		panic(errLis)
	}

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := httpSrv.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("HTTP server shutdown error: %v\n", err)
	}
}

func waitForServer(ctx context.Context, url string) error {
	client := resty.New()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := client.R().Get(url)
			if err == nil && resp.StatusCode() == 200 {
				return nil
			}
		}
	}
}
