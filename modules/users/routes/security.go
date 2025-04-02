package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/handlers/account/security"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

func SetupAccountSecurityRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	securityGroup := apiGroup.Group("/security")

	securityGroup.Put("/email", middleware.AuthMiddleware(db, responseHandler), security.UpdateEmail(db, responseHandler))
	securityGroup.Get("/verify-email", security.VerifyEmail(db, responseHandler))

	securityGroup.Put("/username", middleware.AuthMiddleware(db, responseHandler), security.UpdateUsername(db, responseHandler))
	securityGroup.Get("/search-usernames", security.SearchUsernames(db, responseHandler))

	securityGroup.Delete("/delete-account", middleware.AuthMiddleware(db, responseHandler), security.DeleteAccount(db, responseHandler))

	securityGroup.Put("/change-password", middleware.AuthMiddleware(db, responseHandler), security.ChangePassword(db, responseHandler))

}
