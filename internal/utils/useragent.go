package utils

import (
	"regexp"
	"strings"

	"github.com/mssola/user_agent"
)

// UserAgentInfo 保存从 User-Agent 解析出的客户端信息。
type UserAgentInfo struct {
	// OperatingSystem 操作系统及版本，如 Windows 10、macOS 10.15.7、iOS 17.2
	OperatingSystem string
	// DeviceType 设备类型：Desktop / Mobile / Tablet，无法识别时为空
	DeviceType string
	// BrowserName 浏览器名称，微信内置浏览器归一化为 WeChat
	BrowserName string
	// BrowserVersion 浏览器版本
	BrowserVersion string
}

var (
	// macOSVersionPattern 匹配 "Intel Mac OS X 10_15_7" 中的版本号
	macOSVersionPattern = regexp.MustCompile(`Mac OS X (\d+)[._](\d+)(?:[._](\d+))?`)
	// iosVersionPattern 匹配 "CPU iPhone OS 17_2 like Mac OS X" / "CPU OS 17_2 like Mac OS X"
	iosVersionPattern = regexp.MustCompile(`OS (\d+)[._](\d+)(?:[._](\d+))?`)
	// weChatVersionPattern 匹配 "MicroMessenger/8.0.44.2502"
	weChatVersionPattern = regexp.MustCompile(`MicroMessenger/([\d.]+)`)
)

// ParseUserAgent 解析 User-Agent。
//
// 以 mssola/user_agent 为基础，补齐它识别不到的三处：
//   - 微信内置浏览器：UA 里的 "Version/4.0" 是内核版本，会被当成浏览器版本，
//     真实版本在 "MicroMessenger/x.y.z"；
//   - Android 平板：mssola 对 Android 一律返回 mobile=true，平板 UA 只是缺少
//     "Mobile" 标记，直接采信会把平板算成手机；
//   - Apple 系统版本串：原始串形如 "CPU iPhone OS 17_2 like Mac OS X"、
//     "Intel Mac OS X 10_15_7"，不可直接展示，这里归一化为 iOS / iPadOS / macOS 加版本。
func ParseUserAgent(raw string) UserAgentInfo {
	ua := user_agent.New(raw)
	browserName, browserVersion := ua.Browser()

	info := UserAgentInfo{
		OperatingSystem: parseOperatingSystem(ua),
		DeviceType:      parseDeviceType(raw, ua),
		BrowserName:     browserName,
		BrowserVersion:  browserVersion,
	}

	// 微信 / 企业微信内置浏览器：MicroMessenger 才是真实浏览器版本
	if match := weChatVersionPattern.FindStringSubmatch(raw); match != nil {
		info.BrowserName = "WeChat"
		info.BrowserVersion = match[1]
	}
	return info
}

// parseOperatingSystem 把 UA 解析出的平台与系统串归一化成「名称 + 版本」。
func parseOperatingSystem(ua *user_agent.UserAgent) string {
	platform := ua.Platform()
	rawOS := ua.OS()

	switch platform {
	case "iPhone", "iPod":
		if version := matchVersion(iosVersionPattern, rawOS); version != "" {
			return "iOS " + version
		}
		return "iOS"
	case "iPad":
		if version := matchVersion(iosVersionPattern, rawOS); version != "" {
			return "iPadOS " + version
		}
		return "iPadOS"
	case "Macintosh":
		if version := matchVersion(macOSVersionPattern, rawOS); version != "" {
			return "macOS " + version
		}
		return "macOS"
	case "Windows":
		// mssola 已把 NT 版本映射为 Windows 7 / Windows 10 等可读名称
		if rawOS != "" {
			return rawOS
		}
		return "Windows"
	}

	// Android 的 platform 是 Linux，系统串形如 "Android 13"，可直接使用
	if strings.HasPrefix(rawOS, "Android") {
		return rawOS
	}
	if platform == "Linux" || platform == "X11" {
		return "Linux"
	}
	if rawOS != "" && rawOS != "Other" {
		return rawOS
	}
	if platform != "" && platform != "Other" {
		return platform
	}
	return ""
}

// parseDeviceType 判断设备类型。
//
// 判定顺序：先看明确的移动端平台，再区分 Android 手机与平板，
// 最后非移动端统一归为 Desktop。
func parseDeviceType(raw string, ua *user_agent.UserAgent) string {
	switch ua.Platform() {
	case "iPhone", "iPod":
		return "Mobile"
	case "iPad", "GalaxyTabS4":
		return "Tablet"
	}

	// 必须先于 ua.Mobile() 判断：mssola 对 Android 一律返回 mobile=true，
	// 平板的 UA 只是缺少 "Mobile" 标记。
	if strings.Contains(raw, "Android") {
		if strings.Contains(raw, "Mobile") {
			return "Mobile"
		}
		return "Tablet"
	}

	if ua.Mobile() {
		return "Mobile"
	}
	if ua.Platform() == "" || ua.Platform() == "Other" {
		return ""
	}
	return "Desktop"
}

// matchVersion 从系统串中提取 "主.次[.修订]" 形式的版本号。
func matchVersion(pattern *regexp.Regexp, s string) string {
	match := pattern.FindStringSubmatch(s)
	if match == nil {
		return ""
	}
	version := match[1] + "." + match[2]
	if len(match) > 3 && match[3] != "" {
		version += "." + match[3]
	}
	return version
}
