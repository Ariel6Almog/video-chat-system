package videochat

import (
	"os/exec"
	"sync"
)

type Room struct {
	SessionID  string
	Publishers map[string]bool
	MixerCmd   *exec.Cmd
	mu         sync.Mutex
}

type Hub struct {
	cfg   Config
	rooms map[string]*Room
	mu    sync.Mutex
}

func NewHub(cfg Config) *Hub {
	return &Hub{
		cfg:   cfg,
		rooms: make(map[string]*Room),
	}
}

func (h *Hub) getOrCreateRoom(sessionID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[sessionID]; ok {
		return r
	}
	r := &Room{
		SessionID:  sessionID,
		Publishers: make(map[string]bool),
	}
	h.rooms[sessionID] = r
	return r
}

func (h *Hub) getRoom(sessionID string) (*Room, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.rooms[sessionID]
	return r, ok
}

func (h *Hub) listRooms() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, 0, len(h.rooms))
	for k := range h.rooms {
		out = append(out, k)
	}
	return out
}

func (r *Room) addPublisher(publisherID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Publishers[publisherID] = true
}

func (r *Room) removePublisher(publisherID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Publishers, publisherID)
}

func (r *Room) listPublishers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.Publishers))
	for id := range r.Publishers {
		out = append(out, id)
	}
	return out
}

func (r *Room) mixerRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.MixerCmd != nil && r.MixerCmd.Process != nil
}
