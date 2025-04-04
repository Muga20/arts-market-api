package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/category"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupCategoryRoutes sets up category-related routes
func SetupCategoryRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	categoryGroup := apiGroup.Group("/categories")
	categoryGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	categoryGroup.Post("/", category.CreateCategoryHandler(db, responseHandler))
	categoryGroup.Put("/:id", category.UpdateCategoryHandler(db, responseHandler))
	categoryGroup.Put("/:id/toggle-active", category.ToggleCategoryActiveHandler(db, responseHandler))
	categoryGroup.Delete("/:id", category.DeleteCategoryHandler(db, responseHandler))
	categoryGroup.Get("/categories/:identifier", category.GetCategoryHandler(db, responseHandler))
	categoryGroup.Get("/", category.GetAllCategoriesHandler(db, responseHandler))
}
