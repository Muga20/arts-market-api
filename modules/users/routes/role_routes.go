package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/roles"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupRoleRoutes sets up role-related routes
func SetupRoleRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	roleGroup := apiGroup.Group("/roles")

	// Apply the AuthMiddleware to all routes under /roles
	roleGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	// Role Management Endpoints
	roleGroup.Post("/", roles.CreateRoleHandler(db, responseHandler))
	roleGroup.Put("/:id", roles.UpdateRoleHandler(db, responseHandler))
	roleGroup.Put("/:id/activate", roles.ActivateRoleHandler(db, responseHandler))
	roleGroup.Put("/:id/deactivate", roles.DeactivateRoleHandler(db, responseHandler))
	roleGroup.Get("/:id", roles.GetRoleByIDHandler(db, responseHandler))
	roleGroup.Get("/", roles.GetAllRolesHandler(db, responseHandler))

	// User Role Assignment Endpoints
	roleGroup.Post("/assign", roles.AssignRole(db, responseHandler))
	roleGroup.Post("/remove", roles.RemoveRole(db, responseHandler))
	roleGroup.Get("/for/:user_id", roles.GetUserRoles(db, responseHandler))
}
