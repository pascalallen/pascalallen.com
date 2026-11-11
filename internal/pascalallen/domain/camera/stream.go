package camera

import (
	"context"
	"errors"
)

var ErrUnavailable = errors.New("camera unavailable")

type Stream interface {
	Subscribe(ctx context.Context) (<-chan []byte, func(), error)
	Available() bool
}
