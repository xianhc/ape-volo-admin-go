package utils

import "strings"

func GetDeviceType(platform string, os string, mobileDetected bool) string {
	if platform == "iPhone" {
		return "Mobile"
	}
	if platform == "iPad" || platform == "GalaxyTabS4" {
		return "Tablet"
	}
	if mobileDetected {
		return "Mobile"
	}
	if platform == "Macintosh" || strings.Contains(platform, "Windows") {
		return "Desktop"
	}
	if platform == "Linux" && os == "Android" && !mobileDetected {
		return "Tablet"
	}
	return ""
}
