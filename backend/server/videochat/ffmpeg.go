package videochat

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os/exec"
)

func fnv32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func portForPublisher(base int, sessionID, publisherID string) int {
	return base + int(fnv32(sessionID+"|"+publisherID)%2000) // 5000-6999
}

func mixPortForSession(base int, sessionID string) int {
	return base + int(fnv32("mix|"+sessionID)%1000) // 7000-7999
}

func tsURLForPublisher(cfg Config, sessionID, publisherID string) string {
	port := portForPublisher(cfg.TSBasePort, sessionID, publisherID)
	if cfg.TSMode == "multicast" {
		return fmt.Sprintf("udp://%s:%d?ttl=1&pkt_size=1316", cfg.TSAddr, port)
	}
	return fmt.Sprintf("udp://127.0.0.1:%d?pkt_size=1316", port)
}

func mixOutURL(cfg Config, sessionID string) string {
	port := mixPortForSession(cfg.MixBasePort, sessionID)
	if cfg.TSMode == "multicast" {
		return fmt.Sprintf("udp://%s:%d?ttl=1&pkt_size=1316", cfg.TSAddr, port)
	}
	return fmt.Sprintf("udp://127.0.0.1:%d?pkt_size=1316", port)
}

func startFFmpegIngest(cfg Config, tsURL string) (*exec.Cmd, io.WriteCloser, context.CancelFunc, error) {
	args := []string{
		"-fflags", "+discardcorrupt",
		"-i", "pipe:0",
		"-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac", "-ar", "48000", "-ac", "2",
		"-f", "mpegts",
		tsURL,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, cfg.FFmpegPath, args...)
	stdin, _ := cmd.StdinPipe()
	cmd.Stdout = io.Discard
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, nil, err
	}
	log.Printf("[ingest] ffmpeg → %s (pid=%d)", tsURL, cmd.Process.Pid)
	return cmd, stdin, cancel, nil
}
