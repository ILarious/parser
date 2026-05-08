package model

import "time"

type VKAccount struct {
	ID             int64
	SocialID       int64
	AccountType    string
	FullName       string
	Username       string
	FollowersCount int
	AvatarURL      string
	Private        bool
	Verified       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
