package hey

import (
	"context"
	"fmt"
	"time"

	"github.com/basecamp/hey-sdk/go/pkg/generated"
)

// PostingsService handles posting-level actions (move, seen, trash, etc.).
type PostingsService struct {
	client *Client
}

// NewPostingsService creates a new PostingsService.
func NewPostingsService(client *Client) *PostingsService {
	return &PostingsService{client: client}
}

// MarkSeen marks one or more postings as seen/read.
func (s *PostingsService) MarkSeen(ctx context.Context, postingIDs []int64) (err error) {
	op := OperationInfo{
		Service: "Postings", Operation: "MarkPostingsSeen",
		ResourceType: "posting", IsMutation: true,
	}
	return s.bulkAction(ctx, op, postingIDs, func(ctx context.Context, ids []int64) error {
		body := generated.MarkPostingsRequestContent{PostingIds: ids}
		resp, err := s.genClient().MarkPostingsSeenWithResponse(ctx, body)
		if err != nil {
			return err
		}
		return CheckResponse(resp.HTTPResponse)
	})
}

// MarkUnseen marks one or more postings as unseen/unread.
func (s *PostingsService) MarkUnseen(ctx context.Context, postingIDs []int64) (err error) {
	op := OperationInfo{
		Service: "Postings", Operation: "MarkPostingsUnseen",
		ResourceType: "posting", IsMutation: true,
	}
	return s.bulkAction(ctx, op, postingIDs, func(ctx context.Context, ids []int64) error {
		body := generated.MarkPostingsRequestContent{PostingIds: ids}
		resp, err := s.genClient().MarkPostingsUnseenWithResponse(ctx, body)
		if err != nil {
			return err
		}
		return CheckResponse(resp.HTTPResponse)
	})
}

// Move moves one or more postings to a box.
func (s *PostingsService) Move(ctx context.Context, postingIDs []int64, boxID int64) error {
	op := OperationInfo{
		Service: "Postings", Operation: "MovePostings",
		ResourceType: "posting", IsMutation: true,
	}
	return s.bulkAction(ctx, op, postingIDs, func(ctx context.Context, ids []int64) error {
		return s.movePostings(ctx, ids, boxID)
	})
}

// Trash moves one or more postings to the trash.
func (s *PostingsService) Trash(ctx context.Context, postingIDs []int64) error {
	op := OperationInfo{
		Service: "Postings", Operation: "TrashPostings",
		ResourceType: "posting", IsMutation: true,
	}
	return s.bulkAction(ctx, op, postingIDs, s.trashPostings)
}

// MoveToFeed moves a posting to The Feed.
func (s *PostingsService) MoveToFeed(ctx context.Context, postingID int64) error {
	return s.singleAction(ctx, "MovePostingToFeed", postingID, func(ctx context.Context, id int64) error {
		return s.movePostingToBoxKind(ctx, id, "feedbox")
	})
}

// MoveToSetAside moves a posting to Set Aside.
func (s *PostingsService) MoveToSetAside(ctx context.Context, postingID int64) error {
	return s.singleAction(ctx, "MovePostingToSetAside", postingID, func(ctx context.Context, id int64) error {
		return s.movePostingToBoxKind(ctx, id, "asidebox")
	})
}

// MoveToReplyLater moves a posting to Reply Later.
func (s *PostingsService) MoveToReplyLater(ctx context.Context, postingID int64) error {
	return s.singleAction(ctx, "MovePostingToReplyLater", postingID, func(ctx context.Context, id int64) error {
		return s.movePostingToBoxKind(ctx, id, "laterbox")
	})
}

// MoveToPaperTrail moves a posting to the Paper Trail.
func (s *PostingsService) MoveToPaperTrail(ctx context.Context, postingID int64) error {
	return s.singleAction(ctx, "MovePostingToPaperTrail", postingID, func(ctx context.Context, id int64) error {
		return s.movePostingToBoxKind(ctx, id, "trailbox")
	})
}

// MoveToTrash moves a posting to the trash.
func (s *PostingsService) MoveToTrash(ctx context.Context, postingID int64) error {
	return s.singleAction(ctx, "MovePostingToTrash", postingID, func(ctx context.Context, id int64) error {
		return s.trashPostings(ctx, []int64{id})
	})
}

// Ignore ignores a posting (stops notifications).
func (s *PostingsService) Ignore(ctx context.Context, postingID int64) error {
	return s.singleAction(ctx, "IgnorePosting", postingID, func(ctx context.Context, id int64) error {
		return s.ignorePostings(ctx, []int64{id})
	})
}

// --- Helpers ---

func (s *PostingsService) genClient() *generated.ClientWithResponses {
	s.client.initGeneratedClient()
	return s.client.gen
}

func (s *PostingsService) movePostingToBoxKind(ctx context.Context, postingID int64, boxKind string) error {
	boxID, err := s.boxIDForKind(ctx, boxKind)
	if err != nil {
		return err
	}
	return s.movePostings(ctx, []int64{postingID}, boxID)
}

func (s *PostingsService) boxIDForKind(ctx context.Context, boxKind string) (int64, error) {
	resp, err := s.genClient().ListBoxesWithResponse(ctx)
	if err != nil {
		return 0, err
	}
	if err := CheckResponse(resp.HTTPResponse); err != nil {
		return 0, err
	}
	if resp.JSON200 != nil {
		for _, box := range *resp.JSON200 {
			if box.Kind == boxKind {
				return box.Id, nil
			}
		}
	}
	return 0, fmt.Errorf("hey: box kind %q not found", boxKind)
}

func (s *PostingsService) movePostings(ctx context.Context, postingIDs []int64, boxID int64) error {
	params := &generated.MovePostingsParams{BoxId: boxID}
	body := generated.MovePostingsRequestContent{PostingIds: postingIDs}
	resp, err := s.genClient().MovePostingsWithResponse(ctx, params, body)
	if err != nil {
		return err
	}
	return CheckResponse(resp.HTTPResponse)
}

func (s *PostingsService) trashPostings(ctx context.Context, postingIDs []int64) error {
	body := generated.TrashPostingsRequestContent{PostingIds: postingIDs}
	resp, err := s.genClient().TrashPostingsWithResponse(ctx, body)
	if err != nil {
		return err
	}
	return CheckResponse(resp.HTTPResponse)
}

func (s *PostingsService) ignorePostings(ctx context.Context, postingIDs []int64) error {
	body := generated.IgnorePostingsRequestContent{PostingIds: postingIDs}
	resp, err := s.genClient().IgnorePostingsWithResponse(ctx, body)
	if err != nil {
		return err
	}
	return CheckResponse(resp.HTTPResponse)
}

func (s *PostingsService) singleAction(ctx context.Context, operation string, postingID int64, fn func(context.Context, int64) error) (err error) {
	op := OperationInfo{
		Service: "Postings", Operation: operation,
		ResourceType: "posting", IsMutation: true, ResourceID: postingID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	return fn(ctx, postingID)
}

func (s *PostingsService) bulkAction(ctx context.Context, op OperationInfo, ids []int64, fn func(context.Context, []int64) error) (err error) {
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	return fn(ctx, ids)
}
