package services

import (
	"blog-api/internal/infrastructure"
	"blog-api/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/olivere/elastic/v7"
)

type PostService struct {
	infra *infrastructure.Manager
	ctx   context.Context
}

func NewPostService(infraManager *infrastructure.Manager) *PostService {
	return &PostService{
		infra: infraManager,
		ctx:   context.Background(),
	}
}

// CreatePost creates a new post with transaction support using GORM
func (s *PostService) CreatePost(req *models.CreatePostRequest) (*models.Post, error) {
	db := s.infra.Database.GetDB()

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create post using GORM
	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		Tags:    req.Tags,
	}

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create post: %v", err)
	}

	// Create activity log
	activityLog := models.ActivityLog{
		Action: "new_post",
		PostID: post.ID,
	}

	if err := tx.Create(&activityLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create activity log: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Index in Elasticsearch asynchronously
	go s.indexPostInElasticsearch(&post)

	return &post, nil
}

// GetPost retrieves a post with Cache-Aside pattern using GORM
func (s *PostService) GetPost(id uint) (*models.Post, error) {
	cacheKey := fmt.Sprintf("post:%d", id)
	redis := s.infra.Cache.GetClient()

	// Try to get from Redis first (Cache-Aside pattern)
	cached, err := redis.Get(s.ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var post models.Post
		if err := json.Unmarshal([]byte(cached), &post); err == nil {
			return &post, nil
		}
		log.Printf("Failed to unmarshal cached post: %v", err)
	}

	// Cache miss - get from database using GORM
	db := s.infra.Database.GetDB()
	var post models.Post

	if err := db.First(&post, id).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %v", err)
	}

	// Cache the result with TTL of 5 minutes
	postJSON, err := json.Marshal(post)
	if err == nil {
		redis.Set(s.ctx, cacheKey, postJSON, 5*time.Minute)
	}

	return &post, nil
}

// UpdatePost updates a post and invalidates cache using GORM
func (s *PostService) UpdatePost(id uint, req *models.UpdatePostRequest) (*models.Post, error) {
	db := s.infra.Database.GetDB()

	// First find the post
	var post models.Post
	if err := db.First(&post, id).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to find post: %v", err)
	}

	// Update the post fields
	post.Title = req.Title
	post.Content = req.Content
	post.Tags = req.Tags

	// Save the changes
	if err := db.Save(&post).Error; err != nil {
		return nil, fmt.Errorf("failed to update post: %v", err)
	}

	// Cache invalidation
	cacheKey := fmt.Sprintf("post:%d", id)
	redis := s.infra.Cache.GetClient()
	redis.Del(s.ctx, cacheKey)

	// Update in Elasticsearch asynchronously
	go s.indexPostInElasticsearch(&post)

	return &post, nil
}

// SearchByTag searches posts by tag using GORM with PostgreSQL array operations
func (s *PostService) SearchByTag(tag string) ([]*models.Post, error) {
	db := s.infra.Database.GetDB()

	var posts []*models.Post

	// Use GORM with raw SQL for PostgreSQL array operations
	err := db.Where("tags @> ARRAY[?]", tag).Order("created_at DESC").Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search posts by tag: %v", err)
	}

	return posts, nil
}

// SearchPosts performs full-text search using Elasticsearch

func (s *PostService) SearchPosts(query string) (*models.SearchResponse, error) {
    es := s.infra.Search.GetClient()                    // ✅ Lấy Elasticsearch client

    searchResult, err := es.Search().                   // ✅ Thực hiện search
        Index("posts").                                 // ✅ Chỉ định index
        Query(elastic.NewMultiMatchQuery(query, "title", "content")). // ✅ Multi-match query
        Size(100).                                      // ✅ Giới hạn 100 kết quả
        Do(s.ctx)                                       // ✅ Execute với context

    if err != nil {                                     // ✅ Error handling
        return nil, fmt.Errorf("failed to search posts: %v", err)
    }

    var posts []models.Post                             // ✅ Khởi tạo slice
    for _, hit := range searchResult.Hits.Hits {       // ✅ Iterate qua results
        var esPost models.ElasticsearchPost             // ✅ Unmarshal struct
        if err := json.Unmarshal(hit.Source, &esPost); err != nil {
            log.Printf("Failed to unmarshal Elasticsearch result: %v", err)
            continue                                    // ✅ Skip lỗi, tiếp tục
        }

        post := models.Post{                            // ✅ Convert sang Post struct
            ID:      esPost.ID,
            Title:   esPost.Title,
            Content: esPost.Content,
            Tags:    pq.StringArray(esPost.Tags),       // ✅ Convert string slice sang pq.StringArray
        }
        posts = append(posts, post)                     // ✅ Add vào slice
    }

    return &models.SearchResponse{                      // ✅ Return response struct
        Posts: posts,
        Total: int(searchResult.Hits.TotalHits.Value),  // ✅ Total hits count
    }, nil
}

// indexPostInElasticsearch indexes or updates a post in Elasticsearch
func (s *PostService) indexPostInElasticsearch(post *models.Post) {
	es := s.infra.Search.GetClient()

	esPost := models.ElasticsearchPost{
		ID:      post.ID,
		Title:   post.Title,
		Content: post.Content,
		Tags:    []string(post.Tags),
	}

	_, err := es.Index().
		Index("posts").
		Id(strconv.FormatUint(uint64(post.ID), 10)).
		BodyJson(esPost).
		Do(s.ctx)

	if err != nil {
		log.Printf("Failed to index post in Elasticsearch: %v", err)
	}
}

