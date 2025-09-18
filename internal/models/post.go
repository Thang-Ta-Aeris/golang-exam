package models

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"size:255;not null" validate:"required,min=1,max=255"`
	Content   string         `json:"content" gorm:"type:text;not null" validate:"required,min=1"`
	Tags      pq.StringArray `json:"tags" gorm:"type:text[]"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (Post) TableName() string {
	return "posts"
}

type ActivityLog struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Action   string    `json:"action" gorm:"size:100;not null"`
	PostID   uint      `json:"post_id" gorm:"not null;index"`
	Post     Post      `json:"post,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	LoggedAt time.Time `json:"logged_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (ActivityLog) TableName() string {
	return "activity_logs"
}

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,min=1,max=255"`
	Content string   `json:"content" validate:"required,min=1"`
	Tags    []string `json:"tags"`
}

type UpdatePostRequest struct {
	Title   string   `json:"title" validate:"required,min=1,max=255"`
	Content string   `json:"content" validate:"required,min=1"`
	Tags    []string `json:"tags"`
}

type SearchResponse struct {
	Posts []Post `json:"posts"`
	Total int    `json:"total"`
}

// ElasticsearchPost represents a post document in Elasticsearch
type ElasticsearchPost struct {
	ID      uint     `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// User represents a user entity
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
