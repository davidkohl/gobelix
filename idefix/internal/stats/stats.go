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
	Category021   int
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
func (s *MessageStats) IncrementCategory(cat asterix.Category) {
	s.TotalMessages++

	switch cat {
	case asterix.Cat021:
		s.Category021++
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
		var cat021Pct, cat048Pct, cat062Pct, cat063Pct, otherPct float64
		if s.TotalMessages > 0 {
			total := float64(s.TotalMessages)
			cat021Pct = float64(s.Category021) / total * 100
			cat048Pct = float64(s.Category048) / total * 100
			cat062Pct = float64(s.Category062) / total * 100
			cat063Pct = float64(s.Category063) / total * 100
			otherPct = float64(s.OtherCategory) / total * 100
		}

		logger.Info("Final Statistics",
			"duration", duration.Round(time.Second).String(),
			"total_messages", s.TotalMessages,
			"cat021", s.Category021,
			"cat021_pct", fmt.Sprintf("%.1f%%", cat021Pct),
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
			"total_messages", s.TotalMessages,
			"cat021", s.Category021,
			"cat048", s.Category048,
			"cat062", s.Category062,
			"cat063", s.Category063,
			"other", s.OtherCategory,
			"rate", fmt.Sprintf("%.1f msg/s", rate),
		)
	}
}
