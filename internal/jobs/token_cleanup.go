package jobs

import (
	"context"
	"log"
	"time"

	"agromart2/internal/services"
)

// TokenCleanupService handles cleanup of expired tokens
type TokenCleanupService struct {
	authService services.AuthService
}

// NewTokenCleanupService creates a new token cleanup service
func NewTokenCleanupService(authService services.AuthService) *TokenCleanupService {
	return &TokenCleanupService{
		authService: authService,
	}
}

// CleanupExpiredTokens removes expired tokens from storage
func (s *TokenCleanupService) CleanupExpiredTokens(ctx context.Context) error {
	log.Println("Starting token cleanup job...")
	
	startTime := time.Now()
	if err := s.authService.CleanupExpiredTokens(ctx); err != nil {
		log.Printf("Token cleanup failed: %v", err)
		return err
	}
	
	duration := time.Since(startTime)
	log.Printf("Token cleanup completed in %v", duration)
	return nil
}

// StartCleanupScheduler starts the token cleanup scheduler
func (s *TokenCleanupService) StartCleanupScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Token cleanup scheduler started with interval: %v", interval)

	for {
		select {
		case <-ticker.C:
			cleanupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			if err := s.CleanupExpiredTokens(cleanupCtx); err != nil {
				log.Printf("Scheduled token cleanup failed: %v", err)
			}
			cancel()
		case <-ctx.Done():
			log.Println("Token cleanup scheduler stopped")
			return
		}
	}
}
