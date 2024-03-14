package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command_handler"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/event"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/listener"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/repository"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/routes"
	"log"
	"os"
)

func main() {
	unitOfWork, err := database.NewGormUnitOfWork()
	if err != nil {
		log.Fatal(err)
	}

	database.Migrate(unitOfWork)

	permissionRepository := repository.NewGormPermissionRepository(unitOfWork)
	roleRepository := repository.NewGormRoleRepository(unitOfWork)
	userRepository := repository.NewGormUserRepository(unitOfWork)

	database.Seed(unitOfWork, permissionRepository, roleRepository, userRepository)

	w := messaging.NewRabbitMQConnection()
	defer w.Close()

	commandBus := messaging.NewRabbitMqCommandBus(w)
	eventDispatcher := messaging.NewRabbitMqEventDispatcher(w)

	commandBus.RegisterHandler(command.RegisterUser{}.CommandName(), command_handler.RegisterUserHandler{UserRepository: userRepository, EventDispatcher: eventDispatcher})
	commandBus.RegisterHandler(command.UpdateUser{}.CommandName(), command_handler.UpdateUserHandler{})
	commandBus.RegisterHandler(command.SendWelcomeEmail{}.CommandName(), command_handler.SendWelcomeEmailHandler{EventDispatcher: eventDispatcher})
	eventDispatcher.RegisterListener(event.UserRegistered{}.EventName(), listener.UserRegistration{CommandBus: commandBus})

	go commandBus.StartConsuming()
	go eventDispatcher.StartConsuming()

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := routes.NewRouter()
	router.Config()
	router.Fileserver()
	router.VideoFeed()
	router.Default()
	router.Auth(userRepository, commandBus)
	router.Temp(userRepository)
	router.Serve(":9990")
}
