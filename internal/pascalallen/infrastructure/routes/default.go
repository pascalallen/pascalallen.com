package routes

import "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"

func (r Router) Default() {
	r.defaultFor(wasmModule)
}

func (r Router) defaultFor(module string) {
	r.engine.NoRoute(action.HandleDefault(assetVersion(module)))
}
