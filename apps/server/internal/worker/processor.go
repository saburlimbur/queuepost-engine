package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Processor handles async task execution. It holds DB + logger for publishing.
type Processor struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewProcessor(db *sql.DB, logger *zap.Logger) *Processor {
	return &Processor{DB: db, Logger: logger}
}

// HandlePublishPostTask is the Asynq handler for TypePublishPost.
// Workflow (Phase 3 -> Phase 4 multi-channel):
//  1. Fetch post by ID (must be PENDING/SCHEDULED)
//  2. Fetch social_account for platform + tokens
//  3. Publish via platform API (currently mocked, replaced with real HTTP in Phase 4)
//  4. UPDATE posts SET status = PUBLISHED/FAILED
func (p *Processor) HandlePublishPostTask(ctx context.Context, t *asynq.Task) error {
	var payload PublishPostPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal PublishPostPayload: %w", asynq.SkipRetry)
	}

	p.Logger.Info("processing publish task", zap.Int("post_id", payload.PostID))

	// 1. Fetch post
	var (
		content         string
		socialAccountID int
		status          string
	)
	err := p.DB.QueryRowContext(ctx,
		`SELECT content, social_account_id, status FROM posts WHERE id = $1`,
		payload.PostID,
	).Scan(&content, &socialAccountID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			p.Logger.Warn("post not found, skipping retry", zap.Int("post_id", payload.PostID))
			return fmt.Errorf("post %d not found: %w", payload.PostID, asynq.SkipRetry)
		}
		return fmt.Errorf("query post %d: %w", payload.PostID, err)
	}

	// Idempotency: if already published, skip
	if status == "PUBLISHED" {
		p.Logger.Info("post already published, skipping", zap.Int("post_id", payload.PostID))
		return nil
	}

	// 2. Fetch social account (platform + tokens — needed for real API call)
	var platform string
	err = p.DB.QueryRowContext(ctx,
		`SELECT platform FROM social_accounts WHERE id = $1`, socialAccountID,
	).Scan(&platform)
	if err != nil {
		// If FK is dangling, mark post FAILED and skip retry
		_ = p.markFailed(ctx, payload.PostID, "social account not found")
		return fmt.Errorf("social_account %d not found: %w", socialAccountID, asynq.SkipRetry)
	}

	// 3. Publish (mock for now — replace with LinkedIn/IG Graph API calls in Phase 4)
	//    Example concurrent pattern for multi-channel:
	//    errCh := make(chan error, 1)
	//    go func() { errCh <- publishToLinkedIn(...) }()
	if err := p.publishToPlatform(ctx, platform, content); err != nil {
		_ = p.markFailed(ctx, payload.PostID, err.Error())
		return fmt.Errorf("publish to %s failed: %w", platform, err)
	}

	// 4. Mark PUBLISHED
	_, err = p.DB.ExecContext(ctx,
		`UPDATE posts SET status = 'PUBLISHED', published_at = NOW(), updated_at = NOW() WHERE id = $1`,
		payload.PostID,
	)
	if err != nil {
		return fmt.Errorf("update post %d to PUBLISHED: %w", payload.PostID, err)
	}

	p.Logger.Info("post published successfully",
		zap.Int("post_id", payload.PostID),
		zap.String("platform", platform),
	)
	return nil
}

// publishToPlatform is a placeholder. In Phase 4 this will do real HTTP calls with Go routines.
func (p *Processor) publishToPlatform(ctx context.Context, platform, content string) error {
	p.Logger.Info("mock publish", zap.String("platform", platform), zap.String("content", content))
	// TODO: implement LinkedIn REST API / Instagram Graph API
	// return errors.New("simulated failure") to test retry
	return nil
}

func (p *Processor) markFailed(ctx context.Context, postID int, msg string) error {
	_, err := p.DB.ExecContext(ctx,
		`UPDATE posts SET status = 'FAILED', error_message = $1, updated_at = NOW() WHERE id = $2`,
		msg, postID,
	)
	return err
}
