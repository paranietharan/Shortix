package service

import (
	"context"
	"fmt"
	"log"
	"shortix-api/internal/dto"
	"shortix-api/internal/model"
	"shortix-api/internal/repository"
	"shortix-api/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type URLService interface {
	CreateURL(ctx context.Context, userID string, req *dto.CreateURLRequest) (*dto.CreateURLResponse, error)
	GetRedirectURL(ctx context.Context, shortCode string, clickData *model.Click) (string, error)
	GetAnalytics(ctx context.Context, urlID string) (*dto.AnalyticsResponse, error)
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

func (s *urlService) CreateURL(ctx context.Context, userID string, req *dto.CreateURLRequest) (*dto.CreateURLResponse, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
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
			return nil, err
		}
		if !available {
			return nil, fmt.Errorf("custom alias already taken")
		}
		url.ShortCode = *req.CustomAlias
	} else {
		// Generate short code with collision handling
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			code, err := utils.GenerateShortCode(6)
			if err != nil {
				return nil, err
			}

			available, err := s.urlRepo.IsAliasAvailable(ctx, code)
			if err != nil {
				return nil, err
			}

			if available {
				url.ShortCode = code
				break
			}

			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to generate unique short code")
			}
		}
	}

	if err := s.urlRepo.Create(ctx, url); err != nil {
		return nil, err
	}

	return &dto.CreateURLResponse{
		ID:          url.ID,
		LongURL:     url.LongURL,
		ShortCode:   url.ShortCode,
		CustomAlias: url.CustomAlias,
		ExpiresAt:   url.ExpiresAt,
		CreatedAt:   url.CreatedAt,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, url.ShortCode),
	}, nil
}

func (s *urlService) GetRedirectURL(ctx context.Context, shortCode string, clickData *model.Click) (string, error) {
	// 1. Check Redis cache
	longURL, err := s.cacheRepo.Get(ctx, "url:"+shortCode)
	if err == nil && longURL != "" {
		// Cache hit
		clickData.URLID = s.extractIDFromCache(longURL) // Optional: store ID in cache value too if needed
		s.TrackClick(clickData)
		return s.extractURLFromCache(longURL), nil
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
	select {
	case s.analyticsCh <- click:
	default:
		log.Println("Analytics queue full, dropping event")
	}
}

func (s *urlService) processAnalytics() {
	for click := range s.analyticsCh {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.clickRepo.Create(ctx, click); err != nil {
			log.Printf("Failed to save analytics: %v", err)
		}
		cancel()
	}
}

func (s *urlService) GetAnalytics(ctx context.Context, urlID string) (*dto.AnalyticsResponse, error) {
	return s.clickRepo.GetAnalytics(ctx, urlID)
}

// Helpers for cache value parsing
func (s *urlService) extractIDFromCache(val string) uuid.UUID {
	var idStr string
	fmt.Sscanf(val, "%s|", &idStr)
	id, _ := uuid.Parse(idStr)
	return id
}

func (s *urlService) extractURLFromCache(val string) string {
	var longURL string
	// Split by |
	for i, char := range val {
		if char == '|' {
			longURL = val[i+1:]
			break
		}
	}
	if longURL == "" {
		return val
	}
	return longURL
}
