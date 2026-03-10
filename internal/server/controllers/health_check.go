package controllers

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/manifest"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

var (
	activeRequests         int32 = 0 // how many requests are active
	activeDispatchRequests int32 = 0 // how many plugin dispatching requests are active
)

func CollectActiveRequests() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		atomic.AddInt32(&activeRequests, 1)
		ctx.Next()
		atomic.AddInt32(&activeRequests, -1)
	}
}

func CollectActiveDispatchRequests() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		atomic.AddInt32(&activeDispatchRequests, 1)
		ctx.Next()
		atomic.AddInt32(&activeDispatchRequests, -1)
	}
}

func HealthCheck(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":                   "ok",
			"pool_status":              routine.FetchRoutineStatus(),
			"version":                  manifest.VersionX,
			"build_time":               manifest.BuildTimeX,
			"platform":                 app.Platform,
			"active_requests":          activeRequests,
			"active_dispatch_requests": activeDispatchRequests,
		})
	}
}

// ReadinessCheck returns 200 only when all installed plugins have been initialized
// and are ready to serve requests. Returns 503 while plugins are still starting up.
// This endpoint is intended for use as a startup/readiness probe so that traffic
// is not routed to the instance during plugin pre-compilation.
func ReadinessCheck(manager *plugin_manager.PluginManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !manager.IsReady() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "initializing",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}
