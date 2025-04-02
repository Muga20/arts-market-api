package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/handlers/engagement"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"

	"gorm.io/gorm"
)

func RegisterEngagementRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	engagementGroup := apiGroup.Group("/engagement")

	engagementGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	engagementGroup.Get("/stats/session-user", engagement.GetUserFollowStatsForAuthUser(db, responseHandler))
	engagementGroup.Post("/follow/:id", engagement.FollowUser(db, responseHandler))
	engagementGroup.Delete("/unfollow/:id", engagement.UnfollowUser(db, responseHandler))
	engagementGroup.Get("/stats/:id", engagement.GetUserFollowStats(db, responseHandler))

	// Block / Unblock
	engagementGroup.Post("/block/:id", engagement.BlockUser(db, responseHandler))
	engagementGroup.Delete("/unblock/:id", engagement.UnblockUser(db, responseHandler))
	engagementGroup.Get("/blocked", engagement.GetBlockedUsers(db, responseHandler))
}
