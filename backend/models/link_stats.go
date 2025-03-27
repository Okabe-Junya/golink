package models

import (
	"time"
)

// LinkStats represents statistics for a link
type LinkStats struct {
	LastClickedAt    time.Time      `json:"last_clicked_at" firestore:"last_clicked_at"`
	CreatedAt        time.Time      `json:"created_at" firestore:"created_at"`
	ReferringSites   map[string]int `json:"referring_sites" firestore:"referring_sites"`
	Browsers         map[string]int `json:"browsers" firestore:"browsers"`
	OperatingSystems map[string]int `json:"operating_systems" firestore:"operating_systems"`
	Countries        map[string]int `json:"countries" firestore:"countries"`
	ClicksByDate     map[string]int `json:"clicks_by_date" firestore:"clicks_by_date"`
	DeviceTypes      map[string]int `json:"device_types" firestore:"device_types"`
	Short            string         `json:"short" firestore:"short"`
	Status           string         `json:"status" firestore:"status"`
	TotalClicks      int            `json:"total_clicks" firestore:"total_clicks"`
	UniqueClicks     int            `json:"unique_clicks" firestore:"unique_clicks"`
}

// NewLinkStats creates a new LinkStats with default values
func NewLinkStats(short string) *LinkStats {
	now := time.Now()
	return &LinkStats{
		Short:            short,
		TotalClicks:      0,
		UniqueClicks:     0,
		ReferringSites:   make(map[string]int),
		Browsers:         make(map[string]int),
		OperatingSystems: make(map[string]int),
		Countries:        make(map[string]int),
		ClicksByDate:     make(map[string]int),
		DeviceTypes:      make(map[string]int),
		LastClickedAt:    time.Time{}, // Zero time
		CreatedAt:        now,
		Status:           "active",
	}
}

// RecordClick records a click on the link
func (s *LinkStats) RecordClick(browser, os, country, referrer, deviceType string) {
	// Update total clicks
	s.TotalClicks++

	// Update unique clicks (in a real implementation this would use IP or user ID)
	// For simplicity, we're incrementing by 1
	s.UniqueClicks++

	// Record the browser
	if browser != "" {
		s.Browsers[browser]++
	}

	// Record the operating system
	if os != "" {
		s.OperatingSystems[os]++
	}

	// Record the country
	if country != "" {
		s.Countries[country]++
	}

	// Record the referring site
	if referrer != "" {
		s.ReferringSites[referrer]++
	}

	// Record the device type
	if deviceType != "" {
		s.DeviceTypes[deviceType]++
	}

	// Record the date
	today := time.Now().Format("2006-01-02")
	s.ClicksByDate[today]++

	// Update last clicked time
	s.LastClickedAt = time.Now()
}

// GetTopReferrers returns the top referring sites
func (s *LinkStats) GetTopReferrers(limit int) map[string]int {
	// In a real implementation, this would return the top N referrers
	// For simplicity, we're just returning the map
	return s.ReferringSites
}

// GetClicksByPeriod returns the clicks grouped by period (day, week, month)
func (s *LinkStats) GetClicksByPeriod(period string) map[string]int {
	// In a real implementation, this would aggregate the clicks by the requested period
	// For simplicity, we're just returning the daily clicks
	return s.ClicksByDate
}

// GetTopCountries returns the top countries by clicks
func (s *LinkStats) GetTopCountries(limit int) map[string]int {
	// In a real implementation, this would return the top N countries
	// For simplicity, we're just returning the map
	return s.Countries
}
