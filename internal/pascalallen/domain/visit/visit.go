package visit

import "context"

type Visit struct {
	Language  string
	UserAgent string
}

type Notifier interface {
	Notify(ctx context.Context, v Visit) error
}
