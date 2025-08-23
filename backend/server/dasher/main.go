package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

type Cfg struct {
	Port      string
	InputMode string
	InputAddr string
	BasePort  int
	DashRoot  string
	FFmpeg    string
}

func load() Cfg {
	_ = godotenv.Load(".env")
	return Cfg{
		Port:      env("PORT", "8090"),
		InputMode: env("INPUT_TS_MODE", "unicast"),
		InputAddr: env("INPUT_TS_ADDR", "127.0.0.1"),
		BasePort:  envInt("TS_BASE_PORT", 7000),
		DashRoot:  env("DASH_ROOT", "./vod"),
		FFmpeg:    env("FFMPEG_PATH", "ffmpeg"),
	}
}

func main() {
	cfg := load()
	_ = os.MkdirAll(cfg.DashRoot, 0755)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })
	r.Static("/dash", cfg.DashRoot)

	r.POST("/api/dash/start/:sessionId", func(c *gin.Context) {
		sid := c.Param("sessionId")
		outDir := filepath.Join(cfg.DashRoot, sid)
		_ = os.MkdirAll(outDir, 0755)
		mpd := filepath.Join(outDir, "manifest.mpd")

		port := cfg.BasePort + (int(fnv32("mix|"+sid)) % 1000)
		var tsIn string
		if cfg.InputMode == "multicast" {
			tsIn = fmt.Sprintf("udp://@%s:%d?overrun_nonfatal=1&fifo_size=50000000", cfg.InputAddr, port)
		} else {
			tsIn = fmt.Sprintf("udp://127.0.0.1:%d?overrun_nonfatal=1&fifo_size=50000000", port)
		}

		args := []string{
			"-i", tsIn,
			"-map", "0:v:0", "-map", "0:a:0?",
			"-c:v", "copy", "-c:a", "aac",
			"-f", "dash",
			"-seg_duration", "2",
			"-use_template", "1",
			"-use_timeline", "1",
			"-streaming", "1",
			"-remove_at_exit", "1",
			"-window_size", "10",
			"-min_seg_duration", "1000000",
			"-init_seg_name", "init-$RepresentationID$.m4s",
			"-media_seg_name", "chunk-$RepresentationID$-$Number$.m4s",
			mpd,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cmd := exec.CommandContext(ctx, cfg.FFmpeg, args...)
		cmd.Stdout = io.Discard
		// cmd.Stderr = io.Discard
		if err := cmd.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		go cmd.Wait()
		c.JSON(http.StatusOK, gin.H{"ok": true, "mpd": fmt.Sprintf("/dash/%s/manifest.mpd", sid)})
	})

	log.Println("DASHER on :" + cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

func fnv32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
