package auth

import "time"

type LoginUserInfo struct {
	UserId          int64     `json:"userId"`
	Account         string    `json:"account"`
	NickName        string    `json:"nickName"`
	DeptId          int64     `json:"deptId"`
	DeptName        string    `json:"deptName"`
	Ip              string    `json:"ip"`
	Address         string    `json:"address"`
	OperatingSystem string    `json:"operatingSystem"`
	DeviceType      string    `json:"deviceType"`
	BrowserName     string    `json:"browserName"`
	Version         string    `json:"version"`
	AccessToken     string    `json:"accessToken"`
	LoginTime       time.Time `json:"loginTime"`
	IsAdmin         bool      `json:"isAdmin"`
}
