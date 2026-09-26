package biz

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrGreetingNotFound = errors.NotFound("GREETING_NOT_FOUND", "greeting not found")
)

// Greeter is a Greeter model.
type Greeter struct {
	ID    int64
	Hello string
}

// GreeterRepo stores and retrieves greeting records.
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
}

// GreeterUsecase is a Greeter usecase.
type GreeterUsecase struct {
	repo GreeterRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(repo GreeterRepo) *GreeterUsecase {
	return &GreeterUsecase{repo: repo}
}

// CreateGreeter creates a Greeter, and returns the new Greeter.
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *Greeter) (*Greeter, error) {
	g.Hello = strings.TrimSpace(g.Hello)
	if utf8.RuneCountInString(g.Hello) == 0 || utf8.RuneCountInString(g.Hello) > 256 {
		return nil, errors.BadRequest("INVALID_NAME", "name must contain 1 to 256 characters")
	}
	return uc.repo.Save(ctx, g)
}

func (uc *GreeterUsecase) GetGreeting(ctx context.Context, id int64) (*Greeter, error) {
	if id <= 0 {
		return nil, errors.BadRequest("INVALID_ID", "id must be positive")
	}
	return uc.repo.FindByID(ctx, id)
}
