package handlers

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/pkg/logs/models"
	logs "github.com/muga20/artsMarket/pkg/logs/service"
	"gorm.io/gorm"
)

var (
	callerCache     = make(map[string]*callerInfo)
	callerCacheLock sync.RWMutex
)

type callerInfo struct {
	file   string
	method string
	line   int
}

type ResponseHandler struct {
	logService *logs.LogService
	env        string // "development" or "production"
}

func NewResponseHandler(db *gorm.DB) *ResponseHandler {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &ResponseHandler{
		logService: logs.NewLogService(db),
		env:        env,
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"` // User-friendly message
	Data    interface{} `json:"data,omitempty"`    // Only for successful responses
}

func (h *ResponseHandler) HandleResponse(c *fiber.Ctx, data interface{}, err error) error {
	// Handle error cases first
	if err != nil {
		return h.handleError(c, err)
	}

	// Handle success case
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func (h *ResponseHandler) handleError(c *fiber.Ctx, err error) error {
	// Log the error with context
	h.logError(c, err)

	// Determine status code and message
	status := fiber.StatusInternalServerError
	message := "Something went wrong"

	switch e := err.(type) {
	case *fiber.Error:
		status = e.Code
		message = e.Message
	default:
		// For non-fiber errors in production, use generic message
		if h.env != "development" {
			message = "An unexpected error occurred"
		} else {
			message = err.Error()
		}
	}

	return c.Status(status).JSON(Response{
		Success: false,
		Message: message,
	})
}

func (h *ResponseHandler) logError(c *fiber.Ctx, err error) {
	if err == nil {
		return
	}

	// Extract all needed values from context BEFORE goroutine
	var ip, userAgent string
	if c != nil {
		ip = c.IP()
		userAgent = c.Get("User-Agent")
	} else {
		ip = "unknown"
		userAgent = "unknown"
	}

	// Get caller info synchronously
	pc, file, line, _ := runtime.Caller(3)
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		// Simplify function name
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
		if dot := strings.LastIndex(funcName, "."); dot >= 0 {
			funcName = funcName[dot+1:]
		}
	}

	// Create error log in goroutine but with all values captured
	go func(ip, userAgent, file, funcName string, line int) {
		errLog := &models.ErrorLog{
			Level:      "error",
			Message:    err.Error(),
			FileName:   filepath.Base(file),
			MethodName: funcName,
			LineNumber: line,
			IPAddress:  ip,
			UserAgent:  userAgent,
			OccurredAt: time.Now(),
			StackTrace: h.getStackTrace(),
		}

		if logErr := h.logService.CreateErrorLog(errLog); logErr != nil {
			log.Printf("Failed to log error: %v", logErr)
		}
	}(ip, userAgent, file, funcName, line)
}

func (h *ResponseHandler) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func (h *ResponseHandler) getCachedCaller() (file, method string, line int) {
	pcs := make([]uintptr, 1)
	if runtime.Callers(3, pcs) == 0 {
		return "unknown", "unknown", 0
	}
	pc := pcs[0]

	callerCacheLock.RLock()
	if info, exists := callerCache[string(pc)]; exists {
		callerCacheLock.RUnlock()
		return info.file, info.method, info.line
	}
	callerCacheLock.RUnlock()

	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	file = filepath.Base(frame.File)

	method = "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		fullName := fn.Name()
		if lastSlash := strings.LastIndex(fullName, "/"); lastSlash >= 0 {
			fullName = fullName[lastSlash+1:]
		}
		if dot := strings.LastIndex(fullName, "."); dot >= 0 {
			method = fullName[dot+1:]
		}
	}

	callerCacheLock.Lock()
	callerCache[string(pc)] = &callerInfo{
		file:   file,
		method: method,
		line:   frame.Line,
	}
	callerCacheLock.Unlock()

	return file, method, frame.Line
}
