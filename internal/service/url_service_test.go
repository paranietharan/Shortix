package service

import (
	"context"
	"github.com/google/uuid"
	"shortix-api/internal/dto"
	"shortix-api/internal/model"
	"shortix-api/internal/repository"
	"strings"
	"testing"
	"time"
)

type fakeURLs struct {
	repository.URLRepository
	url       *model.URL
	collision bool
	deleted   bool
}

func (f *fakeURLs) GetByShortCode(context.Context, string) (*model.URL, error) { return f.url, nil }
func (f *fakeURLs) GetByID(context.Context, string) (*model.URL, error)        { return f.url, nil }
func (f *fakeURLs) ExistsByCustomAlias(context.Context, string) (bool, error) {
	return f.collision, nil
}
func (f *fakeURLs) Delete(context.Context, string) error { f.deleted = true; return nil }

type fakeCache struct{ value, deleted string }

func (f *fakeCache) Get(context.Context, string) (string, error) { return f.value, nil }
func (f *fakeCache) Set(_ context.Context, _ string, value string, _ time.Duration) error {
	f.value = value
	return nil
}
func (f *fakeCache) Delete(_ context.Context, key string) error { f.deleted = key; return nil }

func TestRedirectTracksDatabaseAndCachedClicks(t *testing.T) {
	id := uuid.New()
	cache := &fakeCache{}
	repo := &fakeURLs{url: &model.URL{ID: id, LongURL: "https://example.com/path", ShortCode: "link"}}
	s := &urlService{urlRepo: repo, cacheRepo: cache, analyticsCh: make(chan *model.Click, 2)}
	for range 2 {
		click := &model.Click{}
		got, err := s.GetRedirectURL(context.Background(), "link", click)
		if err != nil || got != "https://example.com/path" {
			t.Fatalf("redirect=%q err=%v", got, err)
		}
		if tracked := <-s.analyticsCh; tracked.URLID != id {
			t.Fatal("click lost its URL identity")
		}
	}
}
func TestExpiredAndMissingLinksDoNotTrackClicks(t *testing.T) {
	expired := time.Now().Add(-time.Hour)
	for _, url := range []*model.URL{nil, {ID: uuid.New(), ExpiresAt: &expired}} {
		s := &urlService{urlRepo: &fakeURLs{url: url}, cacheRepo: &fakeCache{}, analyticsCh: make(chan *model.Click, 1)}
		if _, err := s.GetRedirectURL(context.Background(), "missing", &model.Click{}); err == nil {
			t.Fatal("expected redirect rejection")
		}
		if len(s.analyticsCh) != 0 {
			t.Fatal("rejected redirect counted as a click")
		}
	}
}
func TestCustomAliasCollision(t *testing.T) {
	alias := "existing"
	s := &urlService{urlRepo: &fakeURLs{collision: true}}
	_, created, err := s.CreateURL(context.Background(), uuid.NewString(), &dto.CreateURLRequest{LongURL: "https://example.com", CustomAlias: &alias})
	if created || err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("created=%v err=%v", created, err)
	}
}
func TestDeleteChecksOwnershipAndInvalidatesCache(t *testing.T) {
	owner := uuid.New()
	repo := &fakeURLs{url: &model.URL{ID: uuid.New(), UserID: owner, ShortCode: "owned"}}
	cache := &fakeCache{}
	s := &urlService{urlRepo: repo, cacheRepo: cache}
	if err := s.DeleteURL(context.Background(), repo.url.ID.String(), uuid.NewString(), "USER"); err == nil || repo.deleted {
		t.Fatal("non-owner deleted a link")
	}
	if err := s.DeleteURL(context.Background(), repo.url.ID.String(), owner.String(), "USER"); err != nil {
		t.Fatal(err)
	}
	if !repo.deleted || cache.deleted != "url:owned" {
		t.Fatal("deletion did not invalidate cached redirect")
	}
}
