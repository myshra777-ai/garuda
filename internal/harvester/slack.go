package harvester

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
)

// SlackConfig holds configuration for the Slack harvester.
type SlackConfig struct {
	Token            string   `json:"token"`
	ChannelIDs       []string `json:"channel_ids"`
	DecisionKeywords []string `json:"decision_keywords"`
	MinConfidence    float64  `json:"min_confidence"`
}

// SlackHarvester implements Harvester for Slack.
type SlackHarvester struct {
	config    *SlackConfig
	client    *slack.Client
	store     Store
	extractor *Extractor
}

// NewSlackHarvester creates a new Slack harvester.
func NewSlackHarvester(cfg *SlackConfig, store Store) (*SlackHarvester, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("slack token is required")
	}
	if cfg.MinConfidence == 0 {
		cfg.MinConfidence = 0.7
	}
	if len(cfg.DecisionKeywords) == 0 {
		cfg.DecisionKeywords = []string{
			"decided", "we should", "let's go with", "we'll use",
			"we're going with", "i think we should", "proposal",
			"agreed", "consensus", "approved",
		}
	}

	client := slack.New(cfg.Token)
	return &SlackHarvester{
		config:    cfg,
		client:    client,
		store:     store,
		extractor: NewExtractor(cfg.DecisionKeywords),
	}, nil
}

func (h *SlackHarvester) Name() string {
	return "slack"
}

// Harvest fetches messages from Slack channels and extracts decisions.
func (h *SlackHarvester) Harvest(ctx context.Context, since time.Time) ([]*HarvestedDecision, error) {
	var results []*HarvestedDecision

	channels, err := h.getChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	for _, channel := range channels {
		slog.Info("harvesting slack channel", "channel", channel.Name, "id", channel.ID)
		msgs, err := h.fetchMessages(ctx, channel.ID, since)
		if err != nil {
			slog.Error("failed to fetch messages", "channel", channel.ID, "error", err)
			continue
		}

		for _, msg := range msgs {
			if msg.Text == "" {
				continue
			}

			extracted, confidence := h.extractor.Extract(msg.Text)
			if confidence < h.config.MinConfidence {
				continue
			}

			// For now, we skip thread replies (simplification)
			// TODO: Add thread fetching later

			hd := &HarvestedDecision{
				ID:                uuid.New(),
				SourceType:        "slack",
				SourceID:          channel.ID,
				SourceURL:         fmt.Sprintf("https://slack.com/archives/%s/p%s", channel.ID, strings.Replace(msg.Timestamp, ".", "", 1)),
				RawText:           msg.Text,
				ExtractedDecision: extracted,
				Confidence:        confidence,
				HumanValidated:    false,
				CreatedAt:         time.Now().UTC(),
				UpdatedAt:         time.Now().UTC(),
			}

			if err := h.store.SaveHarvestedDecision(ctx, hd); err != nil {
				slog.Error("failed to save harvested decision", "error", err)
				continue
			}
			results = append(results, hd)
		}
	}
	return results, nil
}

// Watch starts a background goroutine that polls for new messages.
func (h *SlackHarvester) Watch(ctx context.Context, since time.Time, callback func(*HarvestedDecision) error) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			decisions, err := h.Harvest(ctx, since)
			if err != nil {
				slog.Error("watch harvest failed", "error", err)
				continue
			}
			for _, d := range decisions {
				if err := callback(d); err != nil {
					slog.Error("callback failed", "error", err)
				}
			}
			since = time.Now().UTC()
		}
	}
}

// getChannels returns channels to harvest.
func (h *SlackHarvester) getChannels(ctx context.Context) ([]slack.Channel, error) {
	if len(h.config.ChannelIDs) > 0 {
		var channels []slack.Channel
		for _, id := range h.config.ChannelIDs {
			ch, err := h.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
				ChannelID: id,
			})
			if err != nil {
				slog.Warn("failed to get channel info", "channel_id", id, "error", err)
				continue
			}
			channels = append(channels, *ch)
		}
		return channels, nil
	}

	params := slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
	}
	channels, _, err := h.client.GetConversationsContext(ctx, &params)
	return channels, err
}

// fetchMessages retrieves messages from a channel since a given time.
func (h *SlackHarvester) fetchMessages(ctx context.Context, channelID string, since time.Time) ([]slack.Message, error) {
	params := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Oldest:    fmt.Sprintf("%d", since.Unix()),
		Limit:     100,
	}
	history, err := h.client.GetConversationHistoryContext(ctx, &params)
	if err != nil {
		return nil, err
	}
	return history.Messages, nil
}
