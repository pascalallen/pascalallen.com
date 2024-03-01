package routes

import "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"

func (r Router) Default() {
	r.engine.NoRoute(action.HandleDefault())
}
