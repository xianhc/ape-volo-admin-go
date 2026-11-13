package auth

import "time"

type LoginAttempt struct {
	Count     int       `json:"count"`
	IsLocked  bool      `json:"isLocked"`
	LockUntil time.Time `json:"lockUntil"`
}
