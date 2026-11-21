package rant

import (
	"context"
	"sort"
	"time"
)

type TagStat struct {
	Tag           string  `json:"tag"`
	Count         int     `json:"count"`
	TotalSeverity int     `json:"totalSeverity"`
	AvgSeverity   float64 `json:"avgSeverity"`
}

type Metrics struct {
	TotalCount      int            `json:"totalCount"`
	AverageSeverity float64        `json:"averageSeverity"`
	PeakDayOfWeek   int            `json:"peakDayOfWeek"`
	PeakHourOfDay   int            `json:"peakHourOfDay"`
	DailySeverity   map[string]int `json:"dailySeverity"`
	HourlyMatrix    [7][24]int     `json:"hourlyMatrix"`
	TopTags         []TagStat      `json:"topTags"`
}

func (s *Service) GetMetrics(ctx context.Context, days int) (Metrics, error) {
	records, err := s.store.FindAll(ctx)
	if err != nil {
		return Metrics{}, err
	}

	var cutoff time.Time
	if days > 0 {
		cutoff = time.Now().UTC().AddDate(0, 0, -days)
	}

	metrics := Metrics{
		DailySeverity: make(map[string]int),
		TopTags:       []TagStat{},
	}

	var severitySum int
	var dayTotals [7]int
	var hourTotals [24]int
	tagAccumulator := make(map[string]*TagStat)

	for _, rec := range records {
		if !cutoff.IsZero() && rec.CreatedAt.Before(cutoff) {
			continue
		}

		sev := int(rec.Severity)
		localTime := rec.CreatedAt.Local()
		dateKey := localTime.Format("2006-01-02")

		metrics.TotalCount++
		severitySum += sev
		metrics.DailySeverity[dateKey] += sev
		metrics.HourlyMatrix[rec.DayOfWeek][rec.HourOfDay] += sev

		dayTotals[rec.DayOfWeek] += sev
		hourTotals[rec.HourOfDay] += sev

		for _, tag := range rec.Tags {
			stat, exists := tagAccumulator[tag]
			if !exists {
				stat = &TagStat{Tag: tag}
				tagAccumulator[tag] = stat
			}
			stat.Count++
			stat.TotalSeverity += sev
		}
	}

	if metrics.TotalCount > 0 {
		metrics.AverageSeverity = float64(severitySum) / float64(metrics.TotalCount)
	}

	maxDayScore := -1
	for day, score := range dayTotals {
		if score > maxDayScore {
			maxDayScore = score
			metrics.PeakDayOfWeek = day
		}
	}

	maxHourScore := -1
	for hour, score := range hourTotals {
		if score > maxHourScore {
			maxHourScore = score
			metrics.PeakHourOfDay = hour
		}
	}

	for _, stat := range tagAccumulator {
		stat.AvgSeverity = float64(stat.TotalSeverity) / float64(stat.Count)
		metrics.TopTags = append(metrics.TopTags, *stat)
	}

	sort.Slice(metrics.TopTags, func(i, j int) bool {
		if metrics.TopTags[i].TotalSeverity == metrics.TopTags[j].TotalSeverity {
			return metrics.TopTags[i].Count > metrics.TopTags[j].Count
		}
		return metrics.TopTags[i].TotalSeverity > metrics.TopTags[j].TotalSeverity
	})

	return metrics, nil
}
