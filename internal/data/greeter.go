package data

import (
	"context"

	"github.com/itsping999/greeting/internal/biz"
)

type greeterRepo struct{ data *Data }

func NewGreeterRepo(data *Data) biz.GreeterRepo { return &greeterRepo{data: data} }
func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.data.mu.Lock()
	defer r.data.mu.Unlock()
	r.data.nextID++
	saved := *g
	saved.ID = r.data.nextID
	r.data.greeters[saved.ID] = saved
	return &saved, nil
}
func (r *greeterRepo) FindByID(ctx context.Context, id int64) (*biz.Greeter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.data.mu.RLock()
	defer r.data.mu.RUnlock()
	g, ok := r.data.greeters[id]
	if !ok {
		return nil, biz.ErrGreetingNotFound
	}
	return &g, nil
}
