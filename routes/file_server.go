package routes

func (r Router) Fileserver() {
	r.engine.Static("/public", "../public")
}
