package data

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/biz"
)

func TestMemoryRepository(t *testing.T) {
	ctx := context.Background()
	store := NewData(log.NewStdLogger(io.Discard))
	repo := NewGreeterRepo(store, nil)
	saved, err := repo.Save(ctx, &biz.Greeter{Name: "World"})
	if err != nil {
		t.Fatal(err)
	}
	saved.Name = "changed outside store"
	found, err := repo.FindByID(ctx, saved.ID)
	if err != nil || found.Name != "World" {
		t.Fatalf("stored copy: %+v %v", found, err)
	}
	found.Name = "changed returned copy"
	found, err = repo.FindByID(ctx, saved.ID)
	if err != nil || found.Name != "World" {
		t.Fatalf("read copy: %+v %v", found, err)
	}
	if _, err := repo.FindByID(ctx, 999); err != biz.ErrGreetingNotFound {
		t.Fatalf("missing record: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.Save(cancelled, &biz.Greeter{}); err != context.Canceled {
		t.Fatalf("cancelled save: %v", err)
	}
	if _, err := repo.FindByID(cancelled, saved.ID); err != context.Canceled {
		t.Fatalf("cancelled read: %v", err)
	}
}
func TestConcurrentSavesHaveUniqueIDs(t *testing.T) {
	store := NewData(log.NewStdLogger(io.Discard))
	repo := NewGreeterRepo(store, nil)
	ids := make(chan int64, 100)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g, err := repo.Save(context.Background(), &biz.Greeter{Name: "World"})
			if err != nil {
				t.Error(err)
				return
			}
			ids <- g.ID
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[int64]bool)
	for id := range ids {
		if seen[id] || id <= 0 {
			t.Fatalf("duplicate or invalid id %d", id)
		}
		seen[id] = true
		g, err := repo.FindByID(context.Background(), id)
		if err != nil || g.Name != "World" {
			t.Fatalf("saved record %d missing: %+v %v", id, g, err)
		}
	}
	if len(seen) != 100 {
		t.Fatalf("got %d records", len(seen))
	}
}
