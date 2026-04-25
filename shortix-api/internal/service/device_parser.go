package service

import "strings"

func ParseDeviceFromUserAgent(ua string) string {
	s := strings.ToLower(ua)
	browser := "unknown-browser"
	os := "unknown-os"

	switch {
	case strings.Contains(s, "edg/"):
		browser = "Edge"
	case strings.Contains(s, "chrome/"):
		browser = "Chrome"
	case strings.Contains(s, "safari/") && !strings.Contains(s, "chrome/"):
		browser = "Safari"
	case strings.Contains(s, "firefox/"):
		browser = "Firefox"
	}

	switch {
	case strings.Contains(s, "windows"):
		os = "Windows"
	case strings.Contains(s, "mac os") || strings.Contains(s, "macintosh"):
		os = "macOS"
	case strings.Contains(s, "android"):
		os = "Android"
	case strings.Contains(s, "iphone") || strings.Contains(s, "ipad") || strings.Contains(s, "ios"):
		os = "iOS"
	case strings.Contains(s, "linux"):
		os = "Linux"
	}

	return browser + " on " + os
}
