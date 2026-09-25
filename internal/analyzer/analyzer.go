package analyzer

import (
	"context"
	"fmt"
	"sort"

	"github.com/skytr1x/wtfdii/internal/models"
	"github.com/skytr1x/wtfdii/internal/packagemanager"
)

// Analyzer orchestrates dependency analysis
type Analyzer struct {
	manager packagemanager.PackageManager
	config  *models.Config
}

// New creates a new Analyzer instance
func New(projectPath string, config *models.Config) (*Analyzer, error) {
	detector := &packagemanager.Detector{}
	manager, err := detector.DetectPackageManager(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect package manager: %w", err)
	}

	return &Analyzer{
		manager: manager,
		config:  config,
	}, nil
}

// Analyze performs full dependency analysis
func (a *Analyzer) Analyze(ctx context.Context, projectPath string) (*models.DependencyTree, error) {
	tree, err := a.manager.GetDependencies(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	return tree, nil
}

// GetDependencyChain returns why a package is installed
func (a *Analyzer) GetDependencyChain(ctx context.Context, projectPath string, packageName string) ([]string, error) {
	return a.manager.GetDependencyChain(ctx, projectPath, packageName)
}

// GetPackagesBySize returns packages sorted by size (largest first)
func (a *Analyzer) GetPackagesBySize(ctx context.Context, projectPath string, limit int) ([]*models.Package, error) {
	tree, err := a.Analyze(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	// Convert map to slice
	packages := make([]*models.Package, 0, len(tree.Packages))
	for _, pkg := range tree.Packages {
		packages = append(packages, pkg)
	}

	// Sort by size descending
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Size > packages[j].Size
	})

	// Apply limit
	if limit > 0 && limit < len(packages) {
		packages = packages[:limit]
	}

	return packages, nil
}

// GetStalePackages returns packages that haven't been updated in a while
func (a *Analyzer) GetStalePackages(ctx context.Context, projectPath string) ([]*models.StalePackage, error) {
	outdated, err := a.manager.GetOutdatedPackages(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get outdated packages: %w", err)
	}

	var stalePackages []*models.StalePackage
	for name, info := range outdated {
		if a.shouldIgnorePackage(name) {
			continue
		}

		stalePackages = append(stalePackages, &models.StalePackage{
			Name:    name,
			Current: info.Current,
			Latest:  info.Latest,
			Type:    info.Type,
		})
	}

	sort.Slice(stalePackages, func(i, j int) bool {
		return stalePackages[i].Name < stalePackages[j].Name
	})

	return stalePackages, nil
}

func (a *Analyzer) shouldIgnorePackage(name string) bool {
	for _, ignored := range a.config.IgnorePackages {
		if ignored == name {
			return true
		}
	}
	return false
}

// GetSecurityIssues returns security vulnerabilities
func (a *Analyzer) GetSecurityIssues(ctx context.Context, projectPath string) ([]models.SecurityIssue, error) {
	return a.manager.RunAudit(ctx, projectPath)
}

// GetUnusedPackages returns potentially unused dependencies
func (a *Analyzer) GetUnusedPackages(ctx context.Context, projectPath string) ([]*models.Package, error) {
	// TODO: Implement in Phase 3
	return nil, fmt.Errorf("not implemented yet")
}

// PackageManagerName returns the detected package manager name
func (a *Analyzer) PackageManagerName() string {
	return a.manager.Name()
}
