package videochat

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterVideoChatRoutes(r *gin.Engine, h *Hub) {
	api := r.Group("/api/video")
	{
		api.GET("/rooms", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"rooms": h.listRooms()})
		})
		api.GET("/room/:sessionId/state", func(c *gin.Context) {
			sessionID := c.Param("sessionId")
			room, ok := h.getRoom(sessionID)
			if !ok {
				c.JSON(http.StatusOK, gin.H{"sessionId": sessionID, "publishers": []string{}, "mixerRunning": false})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"sessionId":    sessionID,
				"publishers":   room.listPublishers(),
				"mixerRunning": room.mixerRunning(),
			})
		})
		api.POST("/room/:sessionId/mix/start", func(c *gin.Context) {
			sessionID := c.Param("sessionId")
			url, err := h.startMixer(sessionID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"ok": true, "tsOut": url})
		})
		api.POST("/room/:sessionId/mix/stop", func(c *gin.Context) {
			sessionID := c.Param("sessionId")
			if err := h.stopMixer(sessionID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	}

	// WebSocket ingest: /ws/ingest/:sessionId/:publisherId
	r.GET("/ws/ingest/:sessionId/:publisherId", h.wsIngest)
}
