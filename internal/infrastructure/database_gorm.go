package infrastructure

import (
	"blog-api/internal/config"
	"blog-api/internal/models"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseClient wraps GORM operations
type DatabaseClient struct {
	db *gorm.DB
}

// NewDatabaseClient creates a new GORM database client
func NewDatabaseClient(cfg *config.Config) (*DatabaseClient, error) {
	// Build connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port)

	// Configure GORM with logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open GORM connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL with GORM: %v", err)
	}

	// Get underlying sql.DB for configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&models.Post{}, &models.ActivityLog{}, &models.User{}); err != nil {
		log.Printf("⚠️  Warning: Auto-migration failed: %v", err)
	} else {
		log.Printf("✅ Database tables migrated successfully")
	}

	log.Printf("✅ PostgreSQL connected successfully with GORM")

	return &DatabaseClient{
		db: db,
	}, nil
}

// GetDB returns the GORM database instance
func (dc *DatabaseClient) GetDB() *gorm.DB {
	return dc.db
}

// GetSQLDB returns the underlying sql.DB for raw queries if needed
func (dc *DatabaseClient) GetSQLDB() (*sql.DB, error) {
	return dc.db.DB()
}

// Close closes the database connection
func (dc *DatabaseClient) Close() error {
	sqlDB, err := dc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}

	log.Printf("✅ PostgreSQL connection closed")
	return nil
}

// Health checks database health
func (dc *DatabaseClient) Health() bool {
	sqlDB, err := dc.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

// Stats returns database connection statistics
func (dc *DatabaseClient) Stats() map[string]interface{} {
	sqlDB, err := dc.db.DB()
	if err != nil {
		return map[string]interface{}{"status": "error", "error": err.Error()}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}
