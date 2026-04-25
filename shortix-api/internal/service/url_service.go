package service

import (
	"context"
	"fmt"
	"log"
	"shortix-api/internal/dto"
	"shortix-api/internal/model"
	"shortix-api/internal/repository"
	"shortix-api/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type URLService interface {
	CreateURL(ctx context.Context, userID string, req *dto.CreateURLRequest) (*dto.CreateURLResponse, bool, error)
	GetRedirectURL(ctx context.Context, shortCode string, clickData *model.Click) (string, error)
	GetAnalytics(ctx context.Context, urlID string) (*dto.AnalyticsResponse, error)
	DeleteURL(ctx context.Context, urlID string, userID string, role string) error
}

type urlService struct {
	urlRepo      repository.URLRepository
	clickRepo    repository.ClickRepository
	cacheRepo    repository.CacheRepository
	analyticsCh  chan *model.Click
	baseURL      string
}

func NewURLService(
	urlRepo repository.URLRepository,
	clickRepo repository.ClickRepository,
	cacheRepo repository.CacheRepository,
	baseURL string,
) URLService {
	s := &urlService{
		urlRepo:     urlRepo,
		clickRepo:   clickRepo,
		cacheRepo:   cacheRepo,
		analyticsCh: make(chan *model.Click, 1000), // Buffered channel for analytics
		baseURL:     baseURL,
	}

	// Start background worker for analytics
	go s.processAnalytics()

	return s
}

func (s *urlService) CreateURL(ctx context.Context, userID string, req *dto.CreateURLRequest) (*dto.CreateURLResponse, bool, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, false, fmt.Errorf("invalid user id")
	}

	// Deduplication: Check if long URL already exists for this user (only if NO custom alias is requested)
	if req.CustomAlias == nil {
		existing, err := s.urlRepo.GetByLongURLAndUser(ctx, req.LongURL, userID)
		if err == nil && existing != nil {
			return &dto.CreateURLResponse{
				ID:          existing.ID,
				LongURL:     existing.LongURL,
				ShortCode:   existing.ShortCode,
				CustomAlias: existing.CustomAlias,
				ExpiresAt:   existing.ExpiresAt,
				CreatedAt:   existing.CreatedAt,
				ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existing.ShortCode),
			}, false, nil
		}
	}

	url := &model.URL{
		UserID:      uID,
		LongURL:     req.LongURL,
		CustomAlias: req.CustomAlias,
		ExpiresAt:   req.ExpiresAt,
	}

	if req.CustomAlias != nil {
		// Validate alias availability
		available, err := s.urlRepo.IsAliasAvailable(ctx, *req.CustomAlias)
		if err != nil {
			return nil, false, err
		}
		if !available {
			return nil, false, fmt.Errorf("custom alias already taken")
		}
		url.ShortCode = *req.CustomAlias
	} else {
		// Generate short code with collision handling
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			code, err := utils.GenerateShortCode(6)
			if err != nil {
				return nil, false, err
			}

			available, err := s.urlRepo.IsAliasAvailable(ctx, code)
			if err != nil {
				return nil, false, err
			}

			if available {
				url.ShortCode = code
				break
			}

			if i == maxRetries-1 {
				return nil, false, fmt.Errorf("failed to generate unique short code")
			}
		}
	}

	if err := s.urlRepo.Create(ctx, url); err != nil {
		return nil, false, err
	}

	return &dto.CreateURLResponse{
		ID:          url.ID,
		LongURL:     url.LongURL,
		ShortCode:   url.ShortCode,
		CustomAlias: url.CustomAlias,
		ExpiresAt:   url.ExpiresAt,
		CreatedAt:   url.CreatedAt,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, url.ShortCode),
	}, true, nil
}

func (s *urlService) GetRedirectURL(ctx context.Context, shortCode string, clickData *model.Click) (string, error) {
	// 1. Check Redis cache
	cacheVal, err := s.cacheRepo.Get(ctx, "url:"+shortCode)
	if err == nil && cacheVal != "" {
		// Cache hit
		longURL := s.parseCacheValue(cacheVal, clickData)
		log.Printf("Cache hit for %s, tracking click for URL ID: %s", shortCode, clickData.URLID)
		s.TrackClick(clickData)
		return longURL, nil
	}

	// 2. If cache miss, query DB
	url, err := s.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if url == nil {
		return "", fmt.Errorf("link not found")
	}

	// 3. Validate expiry
	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("link expired")
	}

	// 4. Store in Redis
	ttl := 24 * time.Hour
	if url.ExpiresAt != nil {
		ttl = time.Until(*url.ExpiresAt)
	}
	
	// Cache value format: "id|long_url" to easily track analytics even on cache hit
	cacheValue := fmt.Sprintf("%s|%s", url.ID.String(), url.LongURL)
	_ = s.cacheRepo.Set(ctx, "url:"+shortCode, cacheValue, ttl)

	// 5. Track analytics asynchronously
	clickData.URLID = url.ID
	s.TrackClick(clickData)

	return url.LongURL, nil
}

func (s *urlService) TrackClick(click *model.Click) {
	log.Printf("Queuing click analytics for URL ID: %s", click.URLID)
	select {
	case s.analyticsCh <- click:
		log.Println("Click analytics queued successfully")
	default:
		log.Println("Analytics queue full, dropping event")
	}
}

func (s *urlService) processAnalytics() {
	log.Println("Analytics background worker started")
	for click := range s.analyticsCh {
		log.Printf("Processing click analytics for URL ID: %s", click.URLID)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.clickRepo.Create(ctx, click); err != nil {
			log.Printf("CRITICAL: Failed to save analytics to DB: %v", err)
		} else {
			log.Printf("Analytics saved successfully for URL ID: %s (Click ID: %d)", click.URLID, click.ID)
		}
		cancel()
	}
}

func (s *urlService) GetAnalytics(ctx context.Context, urlID string) (*dto.AnalyticsResponse, error) {
	return s.clickRepo.GetAnalytics(ctx, urlID)
}

func (s *urlService) DeleteURL(ctx context.Context, urlID string, userID string, role string) error {
	url, err := s.urlRepo.GetByID(ctx, urlID)
	if err != nil {
		return err
	}
	if url == nil {
		return fmt.Errorf("link not found")
	}

	// Only owner or ADMIN can delete
	if url.UserID.String() != userID && role != "ADMIN" {
		return fmt.Errorf("permission denied")
	}

	if err := s.urlRepo.Delete(ctx, urlID); err != nil {
		return err
	}

	// Invalidate cache
	_ = s.cacheRepo.Delete(ctx, "url:"+url.ShortCode)

	return nil
}

// parseCacheValue extracts ID and long URL from "id|long_url" format
func (s *urlService) parseCacheValue(val string, click *model.Click) string {
	parts := strings.SplitN(val, "|", 2)
	if len(parts) < 2 {
		return val
	}
	id, _ := uuid.Parse(parts[0])
	click.URLID = id
	return parts[1]
}
