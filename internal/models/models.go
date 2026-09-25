package models

import "time"

type Package struct {
	Name             string
	Version          string
	Size             int64
	IsDirect         bool
	IsTransitive     bool
	LastUpdated      time.Time
	DependsOn        []string
	DependedOnBy     []string
	License          string
	SecurityIssues   []SecurityIssue
	InstallationPath string
}

type SecurityIssue struct {
	Severity    string
	Title       string
	Description string
	CVE         string
	URL         string
}

type DependencyTree struct {
	ProjectName         string
	PackageManager      string
	DirectCount         int
	TransitiveCount     int
	TotalCount          int
	TotalSize           int64
	Packages            map[string]*Package
	SecurityIssuesCount SecurityStats
	StalePackages       []*Package
	UnusedPackages      []*Package
	DuplicatePackages   map[string][]string
}

type SecurityStats struct {
	High     int
	Moderate int
	Low      int
	Total    int
}

type Config struct {
	StaleThresholdMonths int `json:"stale_threshold_months"`
	IgnorePackages []string `json:"ignore_packages"`
	CustomRules map[string]interface{} `json:"custom_rules"`
}

type OutdatedInfo struct {
	Current    string
	Latest     string
	Type       string
	Homepage   string
}

type StalePackage struct {
	Name       string
	Current    string
	Latest     string
	AgeMonths  int
	Type       string
}

func DefaultConfig() *Config {
	return &Config{
		StaleThresholdMonths: 24, // 2 years
		IgnorePackages:       []string{},
		CustomRules:          make(map[string]interface{}),
	}
}
