package biz

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	ReasonInvalidName = "INVALID_NAME"

	ErrGreetingNotFound = errors.NotFound("GREETING_NOT_FOUND", "greeting not found")
	ErrInvalidName      = errors.BadRequest(ReasonInvalidName, "name must contain 1 to 256 characters")
)

// Greeter is a Greeter model.
type Greeter struct {
	ID      int64
	Name    string
	Message string
	Source  string
}

// GreeterRepo stores or forwards Greeter domain data.
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
}

// GreeterEventPublisher publishes optional Greeter domain events.
type GreeterEventPublisher interface {
	PublishGreeterCreated(context.Context, *Greeter) error
}

// GreeterUsecase is a Greeter usecase.
type GreeterUsecase struct {
	repo      GreeterRepo
	publisher GreeterEventPublisher
	log       *log.Helper
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(
	repo GreeterRepo,
	publisher GreeterEventPublisher,
	logger log.Logger,
) *GreeterUsecase {
	return &GreeterUsecase{
		repo:      repo,
		publisher: publisher,
		log:       log.NewHelper(log.With(logger, "module", "biz/greeter")),
	}
}

// CreateGreeter normalizes input, persists the greeting through the repo, and
// emits an optional event after the main flow succeeds.
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, name string) (*Greeter, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 256 {
		return nil, ErrInvalidName
	}

	g := &Greeter{
		Name:    name,
		Message: "Hello " + name,
	}
	saved, err := uc.repo.Save(ctx, g)
	if err != nil {
		return nil, err
	}
	if uc.publisher != nil {
		if err = uc.publisher.PublishGreeterCreated(ctx, saved); err != nil {
			uc.log.WithContext(ctx).Warnf("publish greeter created event failed: %v", err)
		}
	}
	return saved, nil
}

func (uc *GreeterUsecase) GetGreeting(ctx context.Context, id int64) (*Greeter, error) {
	if id <= 0 {
		return nil, errors.BadRequest("INVALID_ID", "id must be positive")
	}
	return uc.repo.FindByID(ctx, id)
}
