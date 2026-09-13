package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypePublishPost = "post:publish"
	TypeRefreshToken = "token:refresh"
)

// PublishPostPayload is the JSON payload for TypePublishPost.
// Keep it minimal (only post_id) so the worker always fetches fresh data from DB.
type PublishPostPayload struct {
	PostID int `json:"post_id"`
}

// NewPublishPostTask creates an Asynq task for publishing a post.
// scheduledAt is used to compute ProcessIn delay — if scheduled time is in the past, task runs immediately.
func NewPublishPostTask(postID int, scheduledAt time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(PublishPostPayload{PostID: postID})
	if err != nil {
		return nil, err
	}

	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
		asynq.Retention(24 * time.Hour),
		asynq.Queue("critical"),
		asynq.TaskID(createTaskID(postID)),
	}

	if delay := time.Until(scheduledAt); delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	return asynq.NewTask(TypePublishPost, payload, opts...), nil
}

func createTaskID(postID int) string {
	// deterministic task ID prevents duplicate enqueue for the same post
	return fmt.Sprintf("publish-post-%d", postID)
}
