package main

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/role"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
	"github.com/rabbitmq/amqp091-go"
)

type Container struct {
	DatabaseSession        database.Session
	PermissionRepository   permission.PermissionRepository
	RoleRepository         role.RoleRepository
	UserRepository         user.UserRepository
	DatabaseSeeder         database.Seeder
	MessageQueueConnection *amqp091.Connection
	CommandBus             messaging.CommandBus
	EventDispatcher        messaging.EventDispatcher
}

func NewContainer(
	dbSession database.Session,
	permissionRepo permission.PermissionRepository,
	roleRepo role.RoleRepository,
	userRepo user.UserRepository,
	dbSeeder database.Seeder,
	mqConn *amqp091.Connection,
	commandBus messaging.CommandBus,
	eventDispatcher messaging.EventDispatcher,
) Container {
	return Container{
		DatabaseSession:        dbSession,
		PermissionRepository:   permissionRepo,
		RoleRepository:         roleRepo,
		UserRepository:         userRepo,
		DatabaseSeeder:         dbSeeder,
		MessageQueueConnection: mqConn,
		CommandBus:             commandBus,
		EventDispatcher:        eventDispatcher,
	}
}
