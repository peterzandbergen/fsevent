package main

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestEnv(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "FSEVENT_LOG_FORMAT":
			return "JSON"
		default:
			return ""
		}
	}

	args := []string{"testapp", "/home/peza/DevProjects/podsync"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out bytes.Buffer
	if err := run(ctx, args, getenv, &out, &out); err != nil {
		t.Errorf("oops: %s", err.Error())
	}
	t.Errorf("done\n%s", out.String())
}
