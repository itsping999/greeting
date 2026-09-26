package data

import (
	"context"
	v1 "github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/biz"
)

type greeterRepo struct {
	data       *Data
	grpcClient v1.GreeterClient
}

func NewGreeterRepo(store *Data, client v1.GreeterClient) biz.GreeterRepo {
	return &greeterRepo{data: store, grpcClient: client}
}
func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.grpcClient != nil {
		out, err := r.grpcClient.SayHello(ctx, &v1.HelloRequest{Name: g.Name})
		if err != nil {
			return nil, err
		}
		return &biz.Greeter{ID: out.Id, Name: g.Name, Message: out.Message, Source: "remote"}, nil
	}
	r.data.mu.Lock()
	defer r.data.mu.Unlock()
	r.data.nextID++
	saved := *g
	saved.ID = r.data.nextID
	saved.Source = "local"
	r.data.greeters[saved.ID] = saved
	return &saved, nil
}
func (r *greeterRepo) FindByID(ctx context.Context, id int64) (*biz.Greeter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.grpcClient != nil {
		out, err := r.grpcClient.GetGreeting(ctx, &v1.GetGreetingRequest{Id: id})
		if err != nil {
			return nil, err
		}
		return &biz.Greeter{ID: out.Id, Message: out.Message, Source: "remote"}, nil
	}
	r.data.mu.RLock()
	defer r.data.mu.RUnlock()
	g, ok := r.data.greeters[id]
	if !ok {
		return nil, biz.ErrGreetingNotFound
	}
	return &g, nil
}
