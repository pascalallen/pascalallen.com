package container

import (
	"log/slog"
	"os"

	appmessaging "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	domainfeed "github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/visit"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/feed"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/slack"
)

type Config struct {
	GitHubToken     string
	SlackWebhookURL string
	AppEnv          string
}

type Container struct {
	Logger         *slog.Logger
	CommandBus     appmessaging.CommandBus
	QueryBus       appmessaging.QueryBus
	RepositoryFeed domainfeed.RepositoryFeed
	PackageFeed    domainfeed.PackageFeed
	VisitNotifier  visit.Notifier
}

func NewContainer(
	logger *slog.Logger,
	commandBus appmessaging.CommandBus,
	queryBus appmessaging.QueryBus,
	repositoryFeed domainfeed.RepositoryFeed,
	packageFeed domainfeed.PackageFeed,
	visitNotifier visit.Notifier,
) Container {
	return Container{
		Logger:         logger,
		CommandBus:     commandBus,
		QueryBus:       queryBus,
		RepositoryFeed: repositoryFeed,
		PackageFeed:    packageFeed,
		VisitNotifier:  visitNotifier,
	}
}

func NewConfig() Config {
	return Config{
		GitHubToken:     os.Getenv("GITHUB_TOKEN"),
		SlackWebhookURL: os.Getenv("SLACK_DM_URL"),
		AppEnv:          os.Getenv("APP_ENV"),
	}
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func NewRepositoryFeed(config Config, cache *feed.Cache) domainfeed.RepositoryFeed {
	return feed.NewCachedRepositoryFeed(cache, feed.NewGitHubClient(feed.GitHubReposURL, config.GitHubToken, nil))
}

func NewPackageFeed(cache *feed.Cache) domainfeed.PackageFeed {
	return feed.NewCachedPackageFeed(cache, feed.NewNpmClient(feed.NpmSearchURL, nil))
}

func NewVisitNotifier(config Config) visit.Notifier {
	return slack.NewNotifier(config.SlackWebhookURL, config.AppEnv, nil)
}
