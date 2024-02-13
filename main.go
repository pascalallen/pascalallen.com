package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/database"
	http2 "github.com/pascalallen/pascalallen.com/http"
	"github.com/pascalallen/pascalallen.com/http/api/v1/auth"
	"github.com/pascalallen/pascalallen.com/http/api/v1/temp"
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

	permissionRepository := repository.GormPermissionRepository{
		UnitOfWork: unitOfWork,
	}
	roleRepository := repository.GormRoleRepository{
		UnitOfWork: unitOfWork,
	}
	userRepository := repository.GormUserRepository{
		UnitOfWork: unitOfWork,
	}

	database.Seed(unitOfWork, permissionRepository, roleRepository, userRepository)

	messaging.WorkerTest()

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}
	router.LoadHTMLGlob("templates/*")
	router.Static("/public", "./public")
	router.NoRoute(http2.HandleDefault())

	v1 := router.Group("/api/v1")
	{
		a := v1.Group("/auth")
		{
			//a.POST("/register", auth.HandleRegisterUser(userRepository))
			a.POST("/login", auth.HandleLoginUser(userRepository))
			a.PATCH("/refresh", auth.HandleRefreshTokens(userRepository))
			//a.POST("/request-reset", auth.HandleRequestPasswordReset)
			//a.POST("/reset-password", auth.HandleResetPassword)
		}

		v1.GET(
			"/temp",
			middleware.AuthRequired(userRepository),
			temp.HandleTemp(),
		)
		ch := make(chan string)
		v1.POST(
			"/event-stream",
			middleware.AuthRequired(userRepository),
			temp.HandleEventStreamPost(ch),
		)
		v1.GET(
			"/event-stream",
			middleware.EventStreamHeaders(),
			temp.HandleEventStreamGet(ch),
		)
	}

	log.Fatalf("error running HTTP server: %s\n", router.Run(":9990"))
}
