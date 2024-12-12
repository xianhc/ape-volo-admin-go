package ext

import "time"

type Duration int64

const (
	Nanosecond  Duration = 1
	Microsecond          = 1000 * Nanosecond
	Millisecond          = 1000 * Microsecond
	Second               = 1000 * Millisecond
	Minute               = 60 * Second
	Hour                 = 60 * Minute
)

// GetCurrentTime 当前时间
func GetCurrentTime() time.Time {
	return time.Now().Local()
}

// GetTimeDuration 期间
func GetTimeDuration(num int, duration Duration) time.Duration {
	switch duration {
	case Hour:
		return time.Hour * time.Duration(num)
	case Minute:
		return time.Minute * time.Duration(num)
	case Second:
		return time.Second * time.Duration(num)
	default:
		return time.Second * time.Duration(num)
	}
}
