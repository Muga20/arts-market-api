package helpers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Paginate is a helper function to handle pagination logic
func Paginate(c *fiber.Ctx, db *gorm.DB, model interface{}, pageSize int) (int, int, []interface{}, error) {
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		fmt.Sscanf(pageParam, "%d", &page)
	}
	if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
		fmt.Sscanf(pageSizeParam, "%d", &pageSize)
	}

	var total int64
	offset := (page - 1) * pageSize

	// Get the total number of items
	if err := db.Model(model).Count(&total).Error; err != nil {
		return 0, 0, nil, err
	}

	var items []interface{}
	// Fetch the items with pagination
	if err := db.Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return 0, 0, nil, err
	}

	// Calculate the next and previous pages
	totalPages := int(total) / pageSize
	if total%int64(pageSize) != 0 {
		totalPages++
	}

	return totalPages, page, items, nil
}
