package handlers

import (
	"log"
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
}

func NewResponseHandler(db *gorm.DB) *ResponseHandler {
	return &ResponseHandler{
		logService: logs.NewLogService(db),
	}
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *ResponseHandler) HandleResponse(c *fiber.Ctx, data interface{}, err error) error {
	if err != nil {
		// Get caller info
		file, method, line := h.getCachedCaller()

		// Capture context values before spawning goroutine
		errorDetails := &errorContext{
			message:    err.Error(),
			file:       file,
			method:     method,
			line:       line,
			ip:         c.IP(),
			userAgent:  c.Get("User-Agent"),
			occurredAt: time.Now(),
		}

		// Async logging with captured values
		go h.safeLogError(errorDetails) // Called from HandleResponse

		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(Response{
		Success: true,
		Data:    data,
	})
}

type errorContext struct {
	message    string
	file       string
	method     string
	line       int
	ip         string
	userAgent  string
	occurredAt time.Time
}

func (h *ResponseHandler) safeLogError(ctx *errorContext) {
	errLog := &models.ErrorLog{
		Level:      "error",
		Message:    ctx.message,
		FileName:   ctx.file,
		MethodName: ctx.method,
		LineNumber: ctx.line,
		IPAddress:  ctx.ip,
		UserAgent:  ctx.userAgent,
		OccurredAt: ctx.occurredAt,
	}

	if logErr := h.logService.CreateErrorLog(errLog); logErr != nil {
		log.Printf("Failed to log error: %v (Original error: %v at %s:%s:%d)",
			logErr, ctx.message, ctx.file, ctx.method, ctx.line)
	}
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
