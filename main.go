package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/database"
	"github.com/pascalallen/pascalallen.com/domain/auth/permission"
	"github.com/pascalallen/pascalallen.com/domain/auth/role"
	"github.com/pascalallen/pascalallen.com/domain/auth/user"
	http2 "github.com/pascalallen/pascalallen.com/http"
	"github.com/pascalallen/pascalallen.com/http/api/v1/auth"
	"github.com/pascalallen/pascalallen.com/http/middleware"
	"github.com/pascalallen/pascalallen.com/repository"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

func main() {
	unitOfWork, err := database.NewGormUnitOfWork()
	if err != nil {
		log.Fatal(err)
	}

	migrate(unitOfWork)

	var permissionRepository permission.PermissionRepository = repository.GormPermissionRepository{
		UnitOfWork: unitOfWork,
	}
	var roleRepository role.RoleRepository = repository.GormRoleRepository{
		UnitOfWork: unitOfWork,
	}
	var userRepository user.UserRepository = repository.GormUserRepository{
		UnitOfWork: unitOfWork,
	}

	seed(unitOfWork, permissionRepository, roleRepository, userRepository)

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
			func(c *gin.Context) {
				c.JSON(
					http.StatusOK,
					http2.JSendSuccessResponse[string]{
						Status: "success",
						Data:   "Ok",
					},
				)

				return
			},
		)
	}

	log.Fatalf("error running HTTP server: %s\n", router.Run(":9990"))
}

func migrate(unitOfWork *gorm.DB) {
	if err := unitOfWork.AutoMigrate(&permission.Permission{}, &role.Role{}, &user.User{}); err != nil {
		err := fmt.Errorf("failed to auto migrate database: %s", err)
		log.Fatal(err)
	}
}

func seed(unitOfWork *gorm.DB, permissionRepository permission.PermissionRepository, roleRepository role.RoleRepository, userRepository user.UserRepository) {
	dataSeeder := database.DataSeeder{
		UnitOfWork:           unitOfWork,
		PermissionRepository: permissionRepository,
		RoleRepository:       roleRepository,
		UserRepository:       userRepository,
	}
	if err := dataSeeder.Seed(); err != nil {
		log.Fatal(err)
	}
}
