package cache

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestCacheSetGetJSON(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	cache := New(client)

	type payload struct {
		Name string `json:"name"`
	}

	ctx := context.Background()
	if err := cache.SetJSON(ctx, "test:key", payload{Name: "value"}, time.Minute); err != nil {
		t.Fatalf("SetJSON error: %v", err)
	}

	var out payload
	ok, err := cache.GetJSON(ctx, "test:key", &out)
	if err != nil || !ok {
		t.Fatalf("GetJSON missing: ok=%v err=%v", ok, err)
	}
	if out.Name != "value" {
		t.Fatalf("expected value, got %s", out.Name)
	}
}
