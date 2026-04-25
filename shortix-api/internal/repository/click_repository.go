package repository

import (
	"context"
	"shortix-api/internal/dto"
	"shortix-api/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ClickRepository interface {
	Create(ctx context.Context, click *model.Click) error
	GetAnalytics(ctx context.Context, urlID string) (*dto.AnalyticsResponse, error)
}

type clickRepository struct {
	db *pgxpool.Pool
}

func NewClickRepository(db *pgxpool.Pool) ClickRepository {
	return &clickRepository{
		db: db,
	}
}

func (r *clickRepository) Create(ctx context.Context, click *model.Click) error {
	query := `
	INSERT INTO clicks (url_id, ip_address, user_agent, device, referrer)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, clicked_at;
	`

	return r.db.QueryRow(ctx,
		query,
		click.URLID,
		click.IPAddress,
		click.UserAgent,
		click.Device,
		click.Referrer,
	).Scan(&click.ID, &click.ClickedAt)
}

func (r *clickRepository) GetAnalytics(ctx context.Context, urlID string) (*dto.AnalyticsResponse, error) {
	var response dto.AnalyticsResponse

	// Total Clicks
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM clicks WHERE url_id = $1", urlID).Scan(&response.TotalClicks)
	if err != nil {
		return nil, err
	}

	// Clicks Per Day (last 30 days)
	rows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(clicked_at, 'YYYY-MM-DD') as date, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND clicked_at > NOW() - INTERVAL '30 days'
		GROUP BY date
		ORDER BY date ASC
	`, urlID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cpd dto.ClickPerDay
			if err := rows.Scan(&cpd.Date, &cpd.Count); err == nil {
				response.ClicksPerDay = append(response.ClicksPerDay, cpd)
			}
		}
	}

	// Top Referrers
	rows, err = r.db.Query(ctx, `
		SELECT COALESCE(referrer, 'Direct') as ref, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1
		GROUP BY ref
		ORDER BY count DESC
		LIMIT 5
	`, urlID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rc dto.ReferrerCount
			if err := rows.Scan(&rc.Referrer, &rc.Count); err == nil {
				response.TopReferrers = append(response.TopReferrers, rc)
			}
		}
	}

	// Device Breakdown
	rows, err = r.db.Query(ctx, `
		SELECT COALESCE(device, 'Unknown') as dev, COUNT(*) as count
		FROM clicks
		WHERE url_id = $1
		GROUP BY dev
		ORDER BY count DESC
	`, urlID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dc dto.DeviceCount
			if err := rows.Scan(&dc.Device, &dc.Count); err == nil {
				response.DeviceBreakdown = append(response.DeviceBreakdown, dc)
			}
		}
	}

	return &response, nil
}
