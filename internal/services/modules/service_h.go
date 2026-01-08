package modules

import (
	"encoding/json"
	"fmt"
	"time"

	"test-go/pkg/logger"
	"test-go/pkg/response"
	"test-go/pkg/utils"

	"github.com/labstack/echo/v4"
)

// SimpleStreamGenerator creates automated demo events for streams
type SimpleStreamGenerator struct {
	streamID    string
	broadcaster *utils.EventBroadcaster
	running     bool
	stopChan    chan struct{}
}

func NewSimpleStreamGenerator(streamID string, broadcaster *utils.EventBroadcaster) *SimpleStreamGenerator {
	return &SimpleStreamGenerator{
		streamID:    streamID,
		broadcaster: broadcaster,
		stopChan:    make(chan struct{}),
	}
}

func (sg *SimpleStreamGenerator) Start() {
	if sg.running {
		return
	}
	sg.running = true
	go sg.generateEvents()
}

func (sg *SimpleStreamGenerator) Stop() {
	if !sg.running {
		return
	}
	sg.running = false
	select {
	case sg.stopChan <- struct{}{}:
	default:
		close(sg.stopChan)
	}
}

func (sg *SimpleStreamGenerator) IsRunning() bool {
	return sg.running
}

func (sg *SimpleStreamGenerator) generateEvents() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	events := []struct {
		Type    string
		Message string
		Data    map[string]interface{}
	}{
		{"demo_notification", "Service H notification", map[string]interface{}{"priority": "low"}},
		{"demo_metric", "Metric update", map[string]interface{}{"value": 42}},
		{"demo_alert", "System alert", map[string]interface{}{"level": "info"}},
		{"demo_update", "Data updated", map[string]interface{}{"records": 100}},
	}

	i := 0
	for {
		select {
		case <-sg.stopChan:
			return
		case <-ticker.C:
			event := events[i%len(events)]
			i++

			// Add metadata
			data := event.Data
			if data == nil {
				data = make(map[string]interface{})
			}
			data["timestamp"] = time.Now().Unix()
			data["service"] = "service_h"
			data["demo_id"] = i

			sg.broadcaster.Broadcast(sg.streamID, event.Type, event.Message, data)
		}
	}
}

// ServiceH is a super simple demo of using the broadcast utility
// Shows how easy it is to add event streaming to any service!
type ServiceH struct {
	enabled     bool
	broadcaster *utils.EventBroadcaster
	streams     map[string]*SimpleStreamGenerator
	logger      *logger.Logger
}

func NewServiceH(enabled bool, logger *logger.Logger) *ServiceH {
	service := &ServiceH{
		enabled:     enabled,
		broadcaster: utils.NewEventBroadcaster(),
		streams:     make(map[string]*SimpleStreamGenerator),
		logger:      logger,
	}

	if enabled {
		logger.Info("Service H starting - broadcasting made easy!")
		service.startDemoStreams()
		logger.Info("Service H ready!")
	}

	return service
}

func (s *ServiceH) Name() string  { return "Service H (Broadcast Utility Demo)" }
func (s *ServiceH) Enabled() bool { return s.enabled }
func (s *ServiceH) Endpoints() []string {
	return []string{"/events/stream/{stream_id}", "/events/broadcast", "/events/streams"}
}

func (s *ServiceH) RegisterRoutes(g *echo.Group) {
	events := g.Group("/events")
	events.GET("/stream/:stream_id", s.streamEvents)
	events.POST("/broadcast", s.broadcastEvent)
	events.GET("/streams", s.getActiveStreams)
	events.POST("/stream/:stream_id/start", s.startStream)
	events.POST("/stream/:stream_id/stop", s.stopStream)
}

// =========================================
// HANDLER METHODS - Using Broadcast Utility
// =========================================

func (s *ServiceH) streamEvents(c echo.Context) error {
	streamID := c.Param("stream_id")
	client := s.broadcaster.Subscribe(streamID)
	defer s.broadcaster.Unsubscribe(client.ID)

	// SSE headers
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	// Send connection event
	initialEvent := utils.EventData{
		ID:        "connected",
		Type:      "connection",
		Message:   "Connected to stream: " + streamID,
		Data:      map[string]interface{}{"stream_id": streamID, "service": "service_h"},
		Timestamp: time.Now().Unix(),
		StreamID:  streamID,
	}

	s.sendSSEEvent(c, initialEvent)

	// Listen for events
	for {
		select {
		case event := <-client.Channel:
			if err := s.sendSSEEvent(c, event); err != nil {
				return nil
			}
		case <-c.Request().Context().Done():
			return nil
		}
	}
}

func (s *ServiceH) broadcastEvent(c echo.Context) error {
	type BroadcastRequest struct {
		StreamID string                 `json:"stream_id,omitempty"`
		Type     string                 `json:"type" validate:"required"`
		Message  string                 `json:"message" validate:"required"`
		Data     map[string]interface{} `json:"data,omitempty"`
	}

	var req BroadcastRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.Type == "" || req.Message == "" {
		return response.BadRequest(c, "Type and message are required")
	}

	if req.StreamID == "" {
		s.broadcaster.BroadcastToAll(req.Type, req.Message, req.Data)
		return response.Success(c, nil, "Event broadcasted to all streams")
	} else {
		s.broadcaster.Broadcast(req.StreamID, req.Type, req.Message, req.Data)
		return response.Success(c, nil, fmt.Sprintf("Event broadcasted to stream: %s", req.StreamID))
	}
}

func (s *ServiceH) getActiveStreams(c echo.Context) error {
	activeStreams := s.broadcaster.GetActiveStreams()
	totalClients := s.broadcaster.GetTotalClients()
	streamCount := s.broadcaster.GetStreamCount()

	streamInfo := make(map[string]interface{})
	for streamID, clientCount := range activeStreams {
		streamInfo[streamID] = map[string]interface{}{
			"clients": clientCount,
			"active":  true,
		}
	}

	result := map[string]interface{}{
		"streams":       streamInfo,
		"total_clients": totalClients,
		"stream_count":  streamCount,
		"service":       "service_h",
	}

	return response.Success(c, result, "Active streams retrieved")
}

func (s *ServiceH) startStream(c echo.Context) error {
	streamID := c.Param("stream_id")

	if generator, exists := s.streams[streamID]; exists {
		generator.Start()
		return response.Success(c, nil, fmt.Sprintf("Stream '%s' restarted", streamID))
	}

	generator := NewSimpleStreamGenerator(streamID, s.broadcaster)
	s.streams[streamID] = generator
	generator.Start()

	return response.Created(c, nil, fmt.Sprintf("Stream '%s' created and started", streamID))
}

func (s *ServiceH) stopStream(c echo.Context) error {
	streamID := c.Param("stream_id")

	generator, exists := s.streams[streamID]
	if !exists {
		return response.NotFound(c, fmt.Sprintf("Stream '%s' not found", streamID))
	}

	generator.Stop()
	delete(s.streams, streamID)

	return response.Success(c, nil, fmt.Sprintf("Stream '%s' stopped and removed", streamID))
}

// =========================================
// HELPER METHODS
// =========================================

func (s *ServiceH) sendSSEEvent(c echo.Context, event utils.EventData) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.Response(), "data: %s\n\n", eventJSON)
	if err != nil {
		return err
	}

	c.Response().Flush()
	return nil
}

func (s *ServiceH) startDemoStreams() {
	streams := []string{"demo-notifications", "demo-metrics", "demo-alerts"}

	for _, streamID := range streams {
		generator := NewSimpleStreamGenerator(streamID, s.broadcaster)
		s.streams[streamID] = generator
		generator.Start()
	}
}
