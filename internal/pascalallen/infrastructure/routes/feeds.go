package routes

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
)

func (r Router) Feeds(queryBus messaging.QueryBus, commandBus messaging.CommandBus) {
	v := r.engine.Group(v1)
	{
		v.GET("/repos", action.HandleListRepositories(queryBus))
		v.HEAD("/repos", action.HandleListRepositories(queryBus))
		v.GET("/packages", action.HandleListPackages(queryBus))
		v.HEAD("/packages", action.HandleListPackages(queryBus))
		v.POST("/visits", action.HandleRecordVisit(commandBus))
	}
}
