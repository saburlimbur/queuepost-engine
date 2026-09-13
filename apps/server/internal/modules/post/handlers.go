package post

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"outpost-engine/internal/models"
	"outpost-engine/internal/worker"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type PostHandler struct {
	DB          *sql.DB
	AsynqClient *asynq.Client
}

func NewPostHandler(db *sql.DB, asynqClient *asynq.Client) *PostHandler {
	return &PostHandler{
		DB:          db,
		AsynqClient: asynqClient,
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req models.Post
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// user_id from JWT (middleware.AuthMiddleware), fallback to body for backward compat
	if uid, exists := c.Get("user_id"); exists {
		req.UserID = uid.(int)
	}

	query := `
		INSERT INTO posts (user_id, social_account_id, content, media_urls, status, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'PENDING', $5, NOW(), NOW())
		RETURNING id, status, created_at
	`

	var id int
	var status string
	var createdAt time.Time

	// media_urls is JSONB in Postgres (packages/db/migrations/001_initial_schema.sql:24)
	mediaJSON, _ := json.Marshal(req.MediaUrls)
	if mediaJSON == nil {
		mediaJSON = []byte("[]")
	}

	err := h.DB.QueryRow(
		query,
		req.UserID,
		req.SocialAccountID,
		req.Content,
		string(mediaJSON),
		req.ScheduledAt,
	).Scan(&id, &status, &createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post: " + err.Error()})
		return
	}

	// Enqueue async publish task (non-blocking, Redis + Asynq)
	if h.AsynqClient != nil {
		task, err := worker.NewPublishPostTask(id, req.ScheduledAt)
		if err == nil {
			if info, err := h.AsynqClient.Enqueue(task); err != nil {
				// log but don't fail the request — post is still PENDING and can be retried
				// use c.Request.Context() if you have logger in context
				_ = info
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Post scheduled successfully",
		"post_id":    id,
		"status":     status,
		"created_at": createdAt,
	})
}

func (h *PostHandler) GetPosts(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, user_id, social_account_id, content, status, scheduled_at, created_at FROM posts ORDER BY scheduled_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch posts",
		})
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.SocialAccountID, &p.Content, &p.Status, &p.ScheduledAt, &p.CreatedAt); err != nil {
			continue
		}
		posts = append(posts, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}
