package routes

import "github.com/pascalallen/pascalallen.com/http/action"

func (r Router) Default() {
	r.engine.NoRoute(action.HandleDefault())
}
