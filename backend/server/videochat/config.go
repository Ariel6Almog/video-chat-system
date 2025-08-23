package videochat

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret       string
	AllowOrigin     string
	FFmpegPath      string
	TSMode          string
	TSAddr          string
	TSBasePort      int
	MixBasePort     int
	MaxParticipants int
}

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

func LoadVideoChatConfig() Config {
	_ = godotenv.Load(".env")
	return Config{
		JWTSecret:       env("JWT_SECRET", "dev-secret"),
		AllowOrigin:     env("ALLOW_ORIGIN", "http://localhost:5173"),
		FFmpegPath:      env("FFMPEG_PATH", "ffmpeg"),
		TSMode:          env("TS_MODE", "unicast"),
		TSAddr:          env("TS_ADDR", "239.10.10.1"),
		TSBasePort:      envInt("TS_BASE_PORT", 5000),
		MixBasePort:     envInt("MIX_BASE_PORT", 7000),
		MaxParticipants: envInt("MAX_PARTICIPANTS", 4),
	}
}
