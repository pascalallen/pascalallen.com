package listener

import "github.com/pascalallen/pascalallen.com/event"

type Listener interface {
	Handle(event event.Event) error
}
