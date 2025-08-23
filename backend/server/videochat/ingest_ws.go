package videochat

import (
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) wsIngest(c *gin.Context) {
	sessionID := c.Param("sessionId")
	publisherID := c.Param("publisherId")
	token := c.Query("token")

	if !h.validateJWT(token) {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	room := h.getOrCreateRoom(sessionID)
	room.addPublisher(publisherID)
	defer room.removePublisher(publisherID)

	tsURL := tsURLForPublisher(h.cfg, sessionID, publisherID)

	cmd, stdin, cancel, err := startFFmpegIngest(h.cfg, tsURL)
	if err != nil {
		log.Println("ffmpeg start error:", err)
		return
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		defer func() { _ = stdin.Close() }()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				errc <- err
				return
			}
			if mt != websocket.BinaryMessage {
				continue
			}
			if len(data) == 0 {
				continue
			}
			if _, err := stdin.Write(data); err != nil {
				errc <- err
				return
			}
		}
	}()

	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ping.C:
			_ = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(2*time.Second))
		case e := <-errc:
			if e != nil {
				log.Printf("[ingest] ws closed: %v", e)
			}
			return
		}
	}
}

func (h *Hub) validateJWT(token string) bool {
	sec := h.cfg.JWTSecret
	return subtle.ConstantTimeCompare([]byte(token), []byte(sec)) == 1
}
