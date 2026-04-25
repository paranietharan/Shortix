package dto

type ClickPerDay struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ReferrerCount struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

type DeviceCount struct {
	Device string `json:"device"`
	Count  int64  `json:"count"`
}

type AnalyticsResponse struct {
	TotalClicks     int64           `json:"total_clicks"`
	ClicksPerDay    []ClickPerDay   `json:"clicks_per_day"`
	TopReferrers    []ReferrerCount `json:"top_referrers"`
	DeviceBreakdown []DeviceCount   `json:"device_breakdown"`
}
