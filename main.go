package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/command_handler"
	"github.com/pascalallen/pascalallen.com/database"
	"github.com/pascalallen/pascalallen.com/http/action"
	"github.com/pascalallen/pascalallen.com/http/middleware"
	"github.com/pascalallen/pascalallen.com/messaging"
	"github.com/pascalallen/pascalallen.com/repository"
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

	commandBus := messaging.NewCommandBus(w)
	commandBus.RegisterHandler(command.RegisterUser{}.CommandName(), command_handler.RegisterUserHandler{UserRepository: userRepository})
	go commandBus.StartConsuming()

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.Default()
	if err = router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}
	router.LoadHTMLGlob("templates/*")
	router.Static("/public", "./public")
	router.NoRoute(action.HandleDefault())

	v1 := router.Group("/api/v1")
	{
		a := v1.Group("/auth")
		{
			a.POST("/register", action.HandleRegisterUser(userRepository, *commandBus))
			a.POST("/login", action.HandleLoginUser(userRepository))
			a.PATCH("/refresh", action.HandleRefreshTokens(userRepository))
			//a.POST("/request-reset", auth.HandleRequestPasswordReset)
			//a.POST("/reset-password", auth.HandleResetPassword)
		}

		v1.GET(
			"/temp",
			middleware.AuthRequired(userRepository),
			action.HandleTemp(),
		)
		ch := make(chan string)
		v1.POST(
			"/event-stream",
			middleware.AuthRequired(userRepository),
			action.HandleEventStreamPost(ch),
		)
		v1.GET(
			"/event-stream",
			middleware.EventStreamHeaders(),
			action.HandleEventStreamGet(ch),
		)
	}

	log.Fatalf("error running HTTP server: %s\n", router.Run(":9990"))
}
