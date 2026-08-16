package hey

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/basecamp/hey-sdk/go/pkg/generated"
)

// TopicsService handles topic operations.
type TopicsService struct {
	client *Client
}

// NewTopicsService creates a new TopicsService.
func NewTopicsService(client *Client) *TopicsService {
	return &TopicsService{client: client}
}

// Get returns a specific topic by ID.
func (s *TopicsService) Get(ctx context.Context, topicID int64) (result *generated.Topic, err error) {
	op := OperationInfo{
		Service: "Topics", Operation: "GetTopic",
		ResourceType: "topic", IsMutation: false, ResourceID: topicID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.GetTopicWithResponse(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetEntries returns entries for a specific topic.
func (s *TopicsService) GetEntries(ctx context.Context, topicID int64, params *generated.GetTopicEntriesParams) (result *generated.GetTopicEntriesResponseContent, err error) {
	op := OperationInfo{
		Service: "Topics", Operation: "GetTopicEntries",
		ResourceType: "entry", IsMutation: false, ResourceID: topicID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.GetTopicEntriesWithResponse(ctx, topicID, params)
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetSent returns sent topics.
func (s *TopicsService) GetSent(ctx context.Context, params *generated.GetSentTopicsParams) (result *generated.TopicListResponse, err error) {
	op := OperationInfo{
		Service: "Topics", Operation: "GetSentTopics",
		ResourceType: "topic", IsMutation: false,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.GetSentTopicsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetSpam returns spam topics.
func (s *TopicsService) GetSpam(ctx context.Context, params *generated.GetSpamTopicsParams) (result *generated.TopicListResponse, err error) {
	op := OperationInfo{
		Service: "Topics", Operation: "GetSpamTopics",
		ResourceType: "topic", IsMutation: false,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.GetSpamTopicsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrash returns trash topics.
func (s *TopicsService) GetTrash(ctx context.Context, params *generated.GetTrashTopicsParams) (result *generated.TopicListResponse, err error) {
	op := OperationInfo{
		Service: "Topics", Operation: "GetTrashTopics",
		ResourceType: "topic", IsMutation: false,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.GetTrashTopicsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetEverything returns all topics.
func (s *TopicsService) GetEverything(ctx context.Context, params *generated.GetEverythingTopicsParams) (result *generated.TopicListResponse, err error) {
	op := OperationInfo{
		Service: "Topics", Operation: "GetEverythingTopics",
		ResourceType: "topic", IsMutation: false,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.GetEverythingTopicsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Restore moves a topic back to active mail.
func (s *TopicsService) Restore(ctx context.Context, topicID int64) (err error) {
	if topicID <= 0 {
		return fmt.Errorf("topic ID must be positive")
	}

	op := OperationInfo{
		Service: "Topics", Operation: "RestoreTopic",
		ResourceType: "topic", IsMutation: true, ResourceID: topicID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.RestoreTopicWithResponse(ctx, topicID)
	if err != nil {
		return err
	}
	return checkMutationResponse(resp.HTTPResponse)
}

// ScheduleBubbleUp schedules a topic to return on a date in YYYY-MM-DD format.
// Set waitingOn when the reminder depends on someone else.
func (s *TopicsService) ScheduleBubbleUp(ctx context.Context, topicID int64, date string, waitingOn bool) (err error) {
	if topicID <= 0 {
		return fmt.Errorf("topic ID must be positive")
	}
	if _, parseErr := time.Parse("2006-01-02", date); parseErr != nil {
		return fmt.Errorf("bubble up date must use YYYY-MM-DD: %w", parseErr)
	}

	op := OperationInfo{
		Service: "Topics", Operation: "ScheduleTopicBubbleUp",
		ResourceType: "topic", IsMutation: true, ResourceID: topicID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	var params *generated.ScheduleTopicBubbleUpParams
	if waitingOn {
		params = &generated.ScheduleTopicBubbleUpParams{WaitingOn: true}
	}
	values := url.Values{"date": {date}}

	s.client.initGeneratedClient()
	resp, err := s.client.gen.ScheduleTopicBubbleUpWithBodyWithResponse(
		ctx,
		topicID,
		params,
		"application/x-www-form-urlencoded",
		strings.NewReader(values.Encode()),
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Accept", "*/*")
			return nil
		},
	)
	if err != nil {
		return err
	}
	return checkMutationResponse(resp.HTTPResponse)
}

// CancelBubbleUp removes a topic's scheduled Bubble Up.
func (s *TopicsService) CancelBubbleUp(ctx context.Context, topicID int64) (err error) {
	if topicID <= 0 {
		return fmt.Errorf("topic ID must be positive")
	}

	op := OperationInfo{
		Service: "Topics", Operation: "CancelTopicBubbleUp",
		ResourceType: "topic", IsMutation: true, ResourceID: topicID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.CancelTopicBubbleUpWithResponse(ctx, topicID)
	if err != nil {
		return err
	}
	return checkMutationResponse(resp.HTTPResponse)
}

// BubbleUpNow moves a topic to the top of the Imbox immediately.
func (s *TopicsService) BubbleUpNow(ctx context.Context, topicID int64) (err error) {
	if topicID <= 0 {
		return fmt.Errorf("topic ID must be positive")
	}

	op := OperationInfo{
		Service: "Topics", Operation: "BubbleUpTopicNow",
		ResourceType: "topic", IsMutation: true, ResourceID: topicID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	s.client.initGeneratedClient()
	resp, err := s.client.gen.BubbleUpTopicNowWithResponse(ctx, topicID)
	if err != nil {
		return err
	}
	return checkMutationResponse(resp.HTTPResponse)
}
