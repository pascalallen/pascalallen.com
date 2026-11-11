//go:build wireinject
// +build wireinject

package container

import (
	"github.com/google/wire"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/feed"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
)

func InitializeContainer() Container {
	wire.Build(
		NewContainer,
		NewConfig,
		NewLogger,
		messaging.NewSynchronousCommandBus,
		messaging.NewSynchronousQueryBus,
		feed.NewCache,
		NewRepositoryFeed,
		NewPackageFeed,
		NewVisitNotifier,
	)
	return Container{}
}
