package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/biz"
)

type GreeterService struct {
	v1.UnimplementedGreeterServer
	log *log.Helper
	uc  *biz.GreeterUsecase
}

func NewGreeterService(logger log.Logger, uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{log: log.NewHelper(log.With(logger, "module", "service/greeter")), uc: uc}
}
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.GetName()})
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello, Id: g.ID}, nil
}
func (s *GreeterService) GetGreeting(ctx context.Context, in *v1.GetGreetingRequest) (*v1.HelloReply, error) {
	g, err := s.uc.GetGreeting(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello, Id: g.ID}, nil
}
