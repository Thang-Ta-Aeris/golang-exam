package infrastructure

import (
	"blog-api/internal/config"
	"fmt"
	"log"
)

// Manager quản lý tất cả infrastructure clients
type Manager struct {
	Database *DatabaseClient
	Cache    *CacheClient
	Search   *SearchClient
}

// NewManager tạo và khởi tạo tất cả infrastructure clients
func NewManager(cfg *config.Config) (*Manager, error) {
	log.Println("🔧 Initializing infrastructure clients...")

	// Initialize Database client
	dbClient, err := NewDatabaseClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database client: %v", err)
	}

	// Initialize Cache client
	cacheClient, err := NewCacheClient(cfg)
	if err != nil {
		dbClient.Close() // cleanup on error
		return nil, fmt.Errorf("failed to initialize cache client: %v", err)
	}

	// Initialize Search client
	searchClient, err := NewSearchClient(cfg)
	if err != nil {
		dbClient.Close()  // cleanup on error
		cacheClient.Close() // cleanup on error
		return nil, fmt.Errorf("failed to initialize search client: %v", err)
	}

	// Initialize search indexes
	if err := searchClient.InitializePostsIndex(); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize search indexes: %v", err)
	}

	manager := &Manager{
		Database: dbClient,
		Cache:    cacheClient,
		Search:   searchClient,
	}

	log.Println("✅ All infrastructure clients initialized successfully")
	return manager, nil
}

// Close đóng tất cả connections một cách graceful
func (m *Manager) Close() error {
	log.Println("🛑 Shutting down infrastructure clients...")

	var errors []error

	// Close Search client
	if m.Search != nil {
		if err := m.Search.Close(); err != nil {
			errors = append(errors, fmt.Errorf("search client: %v", err))
		}
	}

	// Close Cache client
	if m.Cache != nil {
		if err := m.Cache.Close(); err != nil {
			errors = append(errors, fmt.Errorf("cache client: %v", err))
		}
	}

	// Close Database client
	if m.Database != nil {
		if err := m.Database.Close(); err != nil {
			errors = append(errors, fmt.Errorf("database client: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors while closing infrastructure clients: %v", errors)
	}

	log.Println("✅ All infrastructure clients shut down successfully")
	return nil
}

// Health trả về trạng thái sức khỏe của tất cả clients
func (m *Manager) Health() map[string]bool {
	health := make(map[string]bool)

	if m.Database != nil {
		health["database"] = m.Database.Health()
	} else {
		health["database"] = false
	}

	if m.Cache != nil {
		health["cache"] = m.Cache.Health()
	} else {
		health["cache"] = false
	}

	if m.Search != nil {
		health["search"] = m.Search.Health()
	} else {
		health["search"] = false
	}

	return health
}

// Stats trả về thống kê của tất cả clients
func (m *Manager) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	if m.Database != nil {
		stats["database"] = m.Database.Stats()
	}

	if m.Cache != nil {
		stats["cache"] = m.Cache.Stats()
	}

	if m.Search != nil {
		stats["search"] = m.Search.Stats()
	}

	return stats
}

// IsHealthy kiểm tra xem tất cả clients có healthy không
func (m *Manager) IsHealthy() bool {
	health := m.Health()
	for _, healthy := range health {
		if !healthy {
			return false
		}
	}
	return true
}
