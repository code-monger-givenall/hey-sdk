package hey

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// ClearancesService handles HEY Screener operations.
type ClearancesService struct {
	client *Client
}

// NewClearancesService creates a new ClearancesService.
func NewClearancesService(client *Client) *ClearancesService {
	return &ClearancesService{client: client}
}

// PendingClearance is an email sender awaiting a Screener decision.
type PendingClearance struct {
	ID           int64  `json:"id"`
	EntryID      int64  `json:"entry_id,omitempty"`
	TopicID      int64  `json:"topic_id,omitempty"`
	Name         string `json:"name,omitempty"`
	EmailAddress string `json:"email_address,omitempty"`
	Subject      string `json:"subject,omitempty"`
	FeedBoxID    int64  `json:"feed_box_id,omitempty"`
	TrailBoxID   int64  `json:"trail_box_id,omitempty"`
}

// List returns pending Screener clearances.
func (s *ClearancesService) List(ctx context.Context) (result []PendingClearance, err error) {
	op := OperationInfo{
		Service: "Clearances", Operation: "GetClearances",
		ResourceType: "clearance", IsMutation: false,
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
	resp, err := s.client.gen.GetClearancesWithResponse(ctx, func(_ context.Context, req *http.Request) error {
		req.Header.Set("Accept", "text/html")
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = CheckResponse(resp.HTTPResponse); err != nil {
		return nil, err
	}
	return parseClearancesHTML(string(resp.Body))
}

// Approve screens a sender in. A zero designationBoxID uses the Imbox.
func (s *ClearancesService) Approve(ctx context.Context, clearanceID, designationBoxID int64) error {
	if designationBoxID < 0 {
		return fmt.Errorf("designation box ID must not be negative")
	}
	return s.update(ctx, clearanceID, "approved", designationBoxID)
}

// Deny screens a sender out.
func (s *ClearancesService) Deny(ctx context.Context, clearanceID int64) error {
	return s.update(ctx, clearanceID, "denied", 0)
}

func (s *ClearancesService) update(ctx context.Context, clearanceID int64, status string, designationBoxID int64) (err error) {
	if clearanceID <= 0 {
		return fmt.Errorf("clearance ID must be positive")
	}

	op := OperationInfo{
		Service: "Clearances", Operation: "UpdateClearance",
		ResourceType: "clearance", IsMutation: true, ResourceID: clearanceID,
	}
	if gater, ok := s.client.hooks.(GatingHooks); ok {
		if ctx, err = gater.OnOperationGate(ctx, op); err != nil {
			return
		}
	}
	start := time.Now()
	ctx = s.client.hooks.OnOperationStart(ctx, op)
	defer func() { s.client.hooks.OnOperationEnd(ctx, op, err, time.Since(start)) }()

	values := url.Values{"status": {status}}
	if designationBoxID > 0 {
		values.Set("designation_box_id", strconv.FormatInt(designationBoxID, 10))
	}

	s.client.initGeneratedClient()
	resp, err := s.client.gen.UpdateClearanceWithBodyWithResponse(
		ctx,
		clearanceID,
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
	if resp.HTTPResponse != nil && (resp.HTTPResponse.StatusCode == http.StatusFound || resp.HTTPResponse.StatusCode == http.StatusSeeOther) {
		return nil
	}
	return CheckResponse(resp.HTTPResponse)
}

var clearanceIDPattern = regexp.MustCompile(`^clearance_(\d+)$`)
var clearanceEntryPathPattern = regexp.MustCompile(`^/clearances/entries/(\d+)$`)

func parseClearancesHTML(pageHTML string) ([]PendingClearance, error) {
	doc, err := html.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return nil, fmt.Errorf("parse clearances HTML: %w", err)
	}

	var clearances []PendingClearance
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "article" && hasClassName(node, "clearance") {
			if clearance, ok := parseClearanceNode(node); ok {
				clearances = append(clearances, clearance)
			}
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return clearances, nil
}

func parseClearanceNode(node *html.Node) (PendingClearance, bool) {
	match := clearanceIDPattern.FindStringSubmatch(getAttr(node, "id"))
	if match == nil {
		return PendingClearance{}, false
	}
	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return PendingClearance{}, false
	}

	clearance := PendingClearance{ID: id}
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.ElementNode {
			currentID := getAttr(current, "id")
			switch {
			case currentID == fmt.Sprintf("name_clearance_%d", id):
				clearance.Name = normalizedNodeText(current)
			case currentID == fmt.Sprintf("email_clearance_%d", id):
				clearance.EmailAddress = normalizedNodeText(current)
			case hasClassName(current, "clearance__subject"):
				clearance.Subject = normalizedNodeText(current)
			case current.Data == "turbo-frame":
				if entryMatch := clearanceEntryPathPattern.FindStringSubmatch(getAttr(current, "src")); entryMatch != nil {
					clearance.EntryID, _ = strconv.ParseInt(entryMatch[1], 10, 64)
				}
			case current.Data == "input" && getAttr(current, "name") == "reply_to_topic_id":
				clearance.TopicID, _ = strconv.ParseInt(getAttr(current, "value"), 10, 64)
			case current.Data == "form":
				target, boxID := clearanceDesignation(current)
				switch target {
				case "feedboxButton":
					clearance.FeedBoxID = boxID
				case "trailboxButton":
					clearance.TrailBoxID = boxID
				}
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return clearance, true
}

func clearanceDesignation(form *html.Node) (target string, boxID int64) {
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if value := getAttr(node, "data-clearances-target"); value != "" {
				target = value
			}
			if node.Data == "input" && getAttr(node, "name") == "designation_box_id" {
				boxID, _ = strconv.ParseInt(getAttr(node, "value"), 10, 64)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(form)
	return target, boxID
}

func hasClassName(node *html.Node, className string) bool {
	for _, item := range strings.Fields(getAttr(node, "class")) {
		if item == className {
			return true
		}
	}
	return false
}

func normalizedNodeText(node *html.Node) string {
	return strings.Join(strings.Fields(strings.Join(collectTexts(node), " ")), " ")
}
