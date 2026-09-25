package packagemanager

import (
	"context"

	"github.com/skytr1x/wtfdii/internal/models"
)

type PackageManager interface {
	Name() string
	Detect(projectPath string) (bool, error)
	GetDependencies(ctx context.Context, projectPath string) (*models.DependencyTree, error)
	GetPackageInfo(ctx context.Context, projectPath string, packageName string) (*models.Package, error)
	GetDependencyChain(ctx context.Context, projectPath string, packageName string) ([]string, error)
	GetPackageSize(projectPath string, packageName string) (int64, error)
	RunAudit(ctx context.Context, projectPath string) ([]models.SecurityIssue, error)
	GetOutdatedPackages(ctx context.Context, projectPath string) (map[string]*models.OutdatedInfo, error)
}

type Detector struct{}

func (d *Detector) DetectPackageManager(projectPath string) (PackageManager, error) {
	managers := []PackageManager{
		&PnpmAdapter{},
		&YarnAdapter{},
		&NpmAdapter{},
	}

	for _, manager := range managers {
		detected, err := manager.Detect(projectPath)
		if err != nil {
			continue
		}
		if detected {
			return manager, nil
		}
	}

	return &NpmAdapter{}, nil
}
