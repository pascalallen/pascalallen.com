package routes

func (r Router) Fileserver() {
	r.engine.Static("/static", "./web/static") // TODO: Determine that this has been refactored correctly
}
