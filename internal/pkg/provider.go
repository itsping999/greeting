package provider

import (
	"github.com/google/wire"
	"github.com/itsping999/greeting/internal/pkg/dbx"
	"github.com/itsping999/greeting/internal/pkg/discoveryx"
	"github.com/itsping999/greeting/internal/pkg/grpcx"
	"github.com/itsping999/greeting/internal/pkg/messagingx"
)

var ProviderSet = wire.NewSet(
	dbx.ProviderSet,
	discoveryx.ProviderSet,
	grpcx.ProviderSet,
	messagingx.ProviderSet,
)
