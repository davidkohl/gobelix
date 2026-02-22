// internal/stats/stats.go
package stats

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/davidkohl/gobelix/asterix"
)

// MessageStats tracks statistics about processed ASTERIX messages
type MessageStats struct {
	TotalMessages int
	TotalRecords  int
	Category020   int
	Category021   int
	Category023   int
	Category034   int
	Category048   int
	Category062   int
	Category063   int
	OtherCategory int
	StartTime     time.Time
}

// NewMessageStats creates a new MessageStats struct
func NewMessageStats() *MessageStats {
	return &MessageStats{
		StartTime: time.Now(),
	}
}

// IncrementCategory increments the counter for the specified category
func (s *MessageStats) IncrementCategory(cat asterix.Category, recordCount int) {
	s.TotalMessages++
	s.TotalRecords += recordCount

	switch cat {
	case asterix.Cat020:
		s.Category020++
	case asterix.Cat021:
		s.Category021++
	case asterix.Cat023:
		s.Category023++
	case asterix.Cat034:
		s.Category034++
	case asterix.Cat048:
		s.Category048++
	case asterix.Cat062:
		s.Category062++
	case asterix.Cat063:
		s.Category063++
	default:
		s.OtherCategory++
	}
}

// LogStats logs current statistics
func (s *MessageStats) LogStats(logger *slog.Logger, final bool) {
	if s.TotalMessages == 0 {
		return
	}

	duration := time.Since(s.StartTime)

	// Calculate messages per second
	var rate float64
	if duration.Seconds() > 0 {
		rate = float64(s.TotalMessages) / duration.Seconds()
	}

	// For final stats, include percentages
	if final {
		// Calculate percentages
		var cat020Pct, cat021Pct, cat023Pct, cat034Pct, cat048Pct, cat062Pct, cat063Pct, otherPct float64
		if s.TotalMessages > 0 {
			total := float64(s.TotalMessages)
			cat020Pct = float64(s.Category020) / total * 100
			cat021Pct = float64(s.Category021) / total * 100
			cat023Pct = float64(s.Category023) / total * 100
			cat034Pct = float64(s.Category034) / total * 100
			cat048Pct = float64(s.Category048) / total * 100
			cat062Pct = float64(s.Category062) / total * 100
			cat063Pct = float64(s.Category063) / total * 100
			otherPct = float64(s.OtherCategory) / total * 100
		}

		logger.Info("Final Statistics",
			"duration", duration.Round(time.Second).String(),
			"messages", s.TotalMessages,
			"records", s.TotalRecords,
			"cat020", s.Category020,
			"cat020_pct", fmt.Sprintf("%.1f%%", cat020Pct),
			"cat021", s.Category021,
			"cat021_pct", fmt.Sprintf("%.1f%%", cat021Pct),
			"cat023", s.Category023,
			"cat023_pct", fmt.Sprintf("%.1f%%", cat023Pct),
			"cat034", s.Category034,
			"cat034_pct", fmt.Sprintf("%.1f%%", cat034Pct),
			"cat048", s.Category048,
			"cat048_pct", fmt.Sprintf("%.1f%%", cat048Pct),
			"cat062", s.Category062,
			"cat062_pct", fmt.Sprintf("%.1f%%", cat062Pct),
			"cat063", s.Category063,
			"cat063_pct", fmt.Sprintf("%.1f%%", cat063Pct),
			"other", s.OtherCategory,
			"other_pct", fmt.Sprintf("%.1f%%", otherPct),
			"avg_rate", fmt.Sprintf("%.1f msg/s", rate),
		)
	} else {
		logger.Info("Statistics",
			"duration", duration.Round(time.Second).String(),
			"messages", s.TotalMessages,
			"records", s.TotalRecords,
			"cat020", s.Category020,
			"cat021", s.Category021,
			"cat023", s.Category023,
			"cat034", s.Category034,
			"cat048", s.Category048,
			"cat062", s.Category062,
			"cat063", s.Category063,
			"other", s.OtherCategory,
			"rate", fmt.Sprintf("%.1f msg/s", rate),
		)
	}
}
