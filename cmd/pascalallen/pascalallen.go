package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command_handler"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/event"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/listener"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/role"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/routes"
	"os"
)

func main() {
	container := InitializeContainer()
	defer container.MessageQueueConnection.Close()

	configureDatabase(container)

	runConsumers(container)

	configureServer(container)
}

func configureDatabase(container Container) {
	container.DatabaseSession.AutoMigrate(&permission.Permission{}, &role.Role{}, &user.User{})
	container.DatabaseSeeder.Seed()
}

func runConsumers(container Container) {
	commandBus := container.CommandBus
	eventDispatcher := container.EventDispatcher
	userRepository := container.UserRepository

	// command registry
	commandBus.RegisterHandler(command.RegisterUser{}.CommandName(), command_handler.RegisterUserHandler{UserRepository: userRepository, EventDispatcher: eventDispatcher})
	commandBus.RegisterHandler(command.UpdateUser{}.CommandName(), command_handler.UpdateUserHandler{})
	commandBus.RegisterHandler(command.SendWelcomeEmail{}.CommandName(), command_handler.SendWelcomeEmailHandler{EventDispatcher: eventDispatcher})

	// event registry
	eventDispatcher.RegisterListener(event.UserRegistered{}.EventName(), listener.UserRegistration{CommandBus: commandBus})

	go commandBus.StartConsuming()
	go eventDispatcher.StartConsuming()
}

func configureServer(container Container) {
	commandBus := container.CommandBus
	userRepository := container.UserRepository

	gin.SetMode(os.Getenv("GIN_MODE"))

	router := routes.NewRouter()

	router.Config()
	router.Fileserver()
	router.Default()
	router.Auth(userRepository, commandBus)
	router.Temp(userRepository)
	router.Serve(":9990")
}
