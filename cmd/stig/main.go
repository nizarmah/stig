// Package main is the entry point for stig.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/nizarmah/stig/internal/env"
)

func main() {
	e, err := env.NewEnv()
	if err != nil {
		log.Fatalf("failed to create env: %v", err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	browser := rod.New().
		Context(ctx).
		ControlURL(e.BrowserWSURL)
	if err := browser.Connect(); err != nil {
		log.Fatalf("failed to connect to browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: e.GameURL})
	if err != nil {
		log.Fatalf("failed to create page: %v", err)
	}
	defer page.Close()

	page.MustWaitLoad()

	time.Sleep(5 * time.Second)
}
