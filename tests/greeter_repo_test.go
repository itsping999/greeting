package tests

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	v1 "github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/biz"
	"github.com/itsping999/greeting/internal/data"
	"google.golang.org/grpc"
)

func TestGreeterRepoSaveUsesLocalAdapterWhenRemoteDisabled(t *testing.T) {
	repo := data.NewGreeterRepo(data.NewData(log.NewStdLogger(io.Discard)), nil)

	out, err := repo.Save(context.Background(), &biz.Greeter{Name: "alice", Message: "Hello alice"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out == nil || out.Name != "alice" || out.Message != "Hello alice" || out.Source != "local" {
		t.Fatalf("unexpected greeter: %+v", out)
	}
}

func TestGreeterRepoSaveUsesRemoteAdapterWhenConfigured(t *testing.T) {
	client := &fakeGreeterClient{
		reply: &v1.HelloReply{Message: "Hello from remote"},
	}
	repo := data.NewGreeterRepo(data.NewData(log.NewStdLogger(io.Discard)), client)

	out, err := repo.Save(context.Background(), &biz.Greeter{Name: "alice", Message: "Hello alice"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if client.name != "alice" {
		t.Fatalf("expected remote request name alice, got %q", client.name)
	}
	if out == nil || out.Name != "alice" || out.Message != "Hello from remote" || out.Source != "remote" {
		t.Fatalf("unexpected greeter: %+v", out)
	}
}

func TestGreeterRepoSaveReturnsRemoteError(t *testing.T) {
	client := &fakeGreeterClient{err: errors.New("remote failed")}
	repo := data.NewGreeterRepo(data.NewData(log.NewStdLogger(io.Discard)), client)

	_, err := repo.Save(context.Background(), &biz.Greeter{Name: "alice", Message: "Hello alice"})
	if err == nil {
		t.Fatalf("expected remote error")
	}
}

type fakeGreeterClient struct {
	name  string
	reply *v1.HelloReply
	err   error
}

func (f *fakeGreeterClient) SayHello(_ context.Context, in *v1.HelloRequest, _ ...grpc.CallOption) (*v1.HelloReply, error) {
	f.name = in.Name
	return f.reply, f.err
}

func (f *fakeGreeterClient) GetGreeting(_ context.Context, _ *v1.GetGreetingRequest, _ ...grpc.CallOption) (*v1.HelloReply, error) {
	return f.reply, f.err
}

func TestGreeterRepoRetrievesRemoteRecord(t *testing.T) {
	client := &fakeGreeterClient{reply: &v1.HelloReply{Id: 41, Message: "Hello remote"}}
	repo := data.NewGreeterRepo(data.NewData(log.NewStdLogger(io.Discard)), client)
	out, err := repo.FindByID(context.Background(), 41)
	if err != nil || out.ID != 41 || out.Message != "Hello remote" || out.Source != "remote" {
		t.Fatalf("remote query: %+v %v", out, err)
	}
	client.err = errors.New("remote lookup failed")
	if _, err := repo.FindByID(context.Background(), 41); err != client.err {
		t.Fatalf("remote error lost: %v", err)
	}
}
