package routes

import "log"

func (r Router) Config() {
	if err := r.engine.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}

	r.engine.LoadHTMLGlob("web/template/*")
}
