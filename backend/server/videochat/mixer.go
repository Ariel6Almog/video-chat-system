package videochat

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

func (h *Hub) startMixer(sessionID string) (string, error) {
	room := h.getOrCreateRoom(sessionID)

	_ = h.stopMixer(sessionID)

	pubIDs := room.listPublishers()
	if len(pubIDs) > h.cfg.MaxParticipants {
		pubIDs = pubIDs[:h.cfg.MaxParticipants]
	}

	count := h.cfg.MaxParticipants
	inputArgs := make([]string, 0, count*2)
	filterInputs := make([]string, 0, count)
	audioInputs := make([]string, 0, count)

	for idx := 0; idx < count; idx++ {
		if idx < len(pubIDs) {
			url := tsURLForPublisher(h.cfg, sessionID, pubIDs[idx])
			inputArgs = append(inputArgs, "-i", fmt.Sprintf("%s?overrun_nonfatal=1&fifo_size=50000000", url))
			filterInputs = append(filterInputs, fmt.Sprintf("[%d:v]scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2:black[v%d]", idx, idx))
			audioInputs = append(audioInputs, fmt.Sprintf("[%d:a]", idx))
		} else {

			inputArgs = append(inputArgs, "-f", "lavfi", "-i", "color=size=640x360:rate=30:color=black")
			inputArgs = append(inputArgs, "-f", "lavfi", "-i", "anullsrc=channel_layout=stereo:sample_rate=48000")
			vidIdx := len(filterInputs) + len(audioInputs)
			fmt.Printf("%d", vidIdx)
		}
	}

	var vMap []string
	var aMap []string
	inIdx := 0
	for idx := 0; idx < count; idx++ {
		if idx < len(pubIDs) {
			vMap = append(vMap, fmt.Sprintf("[%d:v]", inIdx))
			aMap = append(aMap, fmt.Sprintf("[%d:a]", inIdx))
			inIdx += 1
		} else {
			vMap = append(vMap, fmt.Sprintf("[%d:v]", inIdx))
			inIdx += 1
			aMap = append(aMap, fmt.Sprintf("[%d:a]", inIdx))
			inIdx += 1
		}
	}

	// scale+pad לכל וידאו
	scalePads := make([]string, 0, count)
	for i := 0; i < count; i++ {
		scalePads = append(scalePads, fmt.Sprintf("%sscale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2:black[v%d]", vMap[i], i))
	}

	// xstack layout (2x2 עבור עד 4 משתתפים)
	layout := "0_0|w0_0|0_h0|w0_h0"
	if count == 1 {
		layout = "0_0"
	} else if count == 2 {
		layout = "0_0|w0_0"
	} else if count == 3 {
		layout = "0_0|w0_0|0_h0"
	}

	var fc strings.Builder
	fc.WriteString(strings.Join(scalePads, ";"))
	fc.WriteString(";")
	vin := make([]string, 0, count)
	for i := 0; i < count; i++ {
		vin = append(vin, fmt.Sprintf("[v%d]", i))
	}
	fc.WriteString(strings.Join(vin, ""))
	fc.WriteString(fmt.Sprintf("xstack=inputs=%d:layout=%s[vout];", count, layout))
	fc.WriteString(strings.Join(aMap, ""))
	fc.WriteString(fmt.Sprintf("amix=inputs=%d:normalize=0[aout]", count))

	args := append([]string{}, inputArgs...)
	args = append(args,
		"-filter_complex", fc.String(),
		"-map", "[vout]", "-map", "[aout]",
		"-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-ar", "48000", "-ac", "2",
		"-f", "mpegts",
		mixOutURL(h.cfg, sessionID),
	)

	cmd := exec.Command(h.cfg.FFmpegPath, args...)
	cmd.Stdout = io.Discard
	// cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return "", err
	}
	room.mu.Lock()
	room.MixerCmd = cmd
	room.mu.Unlock()
	log.Printf("[mixer] started (pid=%d) session=%s inputs=%d", cmd.Process.Pid, sessionID, count)
	return mixOutURL(h.cfg, sessionID), nil
}

func (h *Hub) stopMixer(sessionID string) error {
	r, ok := h.getRoom(sessionID)
	if !ok {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.MixerCmd != nil && r.MixerCmd.Process != nil {
		_ = r.MixerCmd.Process.Kill()
		_ = r.MixerCmd.Wait()
		r.MixerCmd = nil
		log.Printf("[mixer] stopped session=%s", sessionID)
	}
	return nil
}
