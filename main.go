package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/command_handler"
	"github.com/pascalallen/pascalallen.com/database"
	"github.com/pascalallen/pascalallen.com/domain/permission"
	"github.com/pascalallen/pascalallen.com/domain/role"
	"github.com/pascalallen/pascalallen.com/domain/user"
	"github.com/pascalallen/pascalallen.com/event"
	"github.com/pascalallen/pascalallen.com/listener"
	"github.com/pascalallen/pascalallen.com/messaging"
	"github.com/pascalallen/pascalallen.com/repository"
	"github.com/pascalallen/pascalallen.com/routes"
	"log"
	"os"
)

func main() {
	unitOfWork, err := database.NewGormUnitOfWork()
	if err != nil {
		log.Fatal(err)
	}

	database.Migrate(unitOfWork)

	var permissionRepository permission.PermissionRepository = repository.NewGormPermissionRepository(unitOfWork)
	var roleRepository role.RoleRepository = repository.NewGormRoleRepository(unitOfWork)
	var userRepository user.UserRepository = repository.NewGormUserRepository(unitOfWork)

	database.Seed(unitOfWork, permissionRepository, roleRepository, userRepository)

	w := messaging.NewRabbitMQConnection()
	defer w.Close()

	commandBus := messaging.NewCommandBus(w)
	eventDispatcher := messaging.NewEventDispatcher(w)

	commandBus.RegisterHandler(command.RegisterUser{}.CommandName(), command_handler.RegisterUserHandler{UserRepository: userRepository, EventDispatcher: *eventDispatcher})
	commandBus.RegisterHandler(command.UpdateUser{}.CommandName(), command_handler.UpdateUserHandler{})
	commandBus.RegisterHandler(command.SendWelcomeEmail{}.CommandName(), command_handler.SendWelcomeEmailHandler{})

	go commandBus.StartConsuming()

	eventDispatcher.RegisterListener(event.UserRegistered{}.EventName(), listener.UserRegistration{CommandBus: *commandBus})

	go eventDispatcher.StartConsuming()

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.Default()
	routes.Config(router)
	routes.Fileserver(router)
	routes.Default(router)
	routes.Auth(router, userRepository, *commandBus)
	routes.Temp(router, userRepository)

	log.Fatalf("error running HTTP server: %s\n", router.Run(":9990"))
}
