package packagemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skytr1x/wtfdii/internal/models"
)

type PnpmAdapter struct{}

func (p *PnpmAdapter) Name() string {
	return "pnpm"
}

func (p *PnpmAdapter) Detect(projectPath string) (bool, error) {
	lockPath := filepath.Join(projectPath, "pnpm-lock.yaml")
	if _, err := os.Stat(lockPath); err == nil {
		return true, nil
	}
	return false, nil
}

func (p *PnpmAdapter) GetDependencies(ctx context.Context, projectPath string) (*models.DependencyTree, error) {
	tree := &models.DependencyTree{
		PackageManager: "pnpm",
		Packages:       make(map[string]*models.Package),
	}

	packageJSON, err := p.readPackageJSON(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	tree.ProjectName = packageJSON.Name
	cmd := exec.CommandContext(ctx, "pnpm", "list", "--json", "--depth", "Infinity", "--long")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to get pnpm list: %w", err)
		}
	}

	var listResult []pnpmListResult
	if err := json.Unmarshal(output, &listResult); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm list output: %w", err)
	}

	directDeps := make(map[string]bool)
	for name := range packageJSON.Dependencies {
		directDeps[name] = true
	}
	for name := range packageJSON.DevDependencies {
		directDeps[name] = true
	}

	if len(listResult) > 0 {
		p.parseDependencies(&listResult[0], tree, directDeps)
	}

	tree.DirectCount = len(directDeps)
	tree.TransitiveCount = len(tree.Packages) - tree.DirectCount
	tree.TotalCount = len(tree.Packages)

	for _, pkg := range tree.Packages {
		tree.TotalSize += pkg.Size
	}

	return tree, nil
}

func (p *PnpmAdapter) parseDependencies(node *pnpmListResult, tree *models.DependencyTree, directDeps map[string]bool) {
	for name, dep := range node.Dependencies {
		pkg := &models.Package{
			Name:         name,
			Version:      dep.Version,
			IsDirect:     directDeps[name],
			IsTransitive: !directDeps[name],
			DependsOn:    []string{},
			DependedOnBy: []string{},
		}

		if dep.Path != "" {
			if size, err := p.calculatePackageSize(dep.Path); err == nil {
				pkg.Size = size
			}
		}

		tree.Packages[name] = pkg

		if len(dep.Dependencies) > 0 {
			p.parseDependencies(&pnpmListResult{Dependencies: dep.Dependencies}, tree, directDeps)
		}
	}
}

func (p *PnpmAdapter) GetPackageInfo(ctx context.Context, projectPath string, packageName string) (*models.Package, error) {
	tree, err := p.GetDependencies(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	pkg, exists := tree.Packages[packageName]
	if !exists {
		return nil, fmt.Errorf("package %s not found", packageName)
	}

	return pkg, nil
}

func (p *PnpmAdapter) GetDependencyChain(ctx context.Context, projectPath string, packageName string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "pnpm", "why", packageName)
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency chain: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var chain []string

	for _, line := range lines {
		if strings.Contains(line, packageName) || strings.Contains(line, "dependencies:") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				chain = append(chain, parts[0])
			}
		}
	}

	return chain, nil
}

func (p *PnpmAdapter) GetPackageSize(projectPath string, packageName string) (int64, error) {
	nodeModulesPath := filepath.Join(projectPath, "node_modules", packageName)
	return p.calculatePackageSize(nodeModulesPath)
}

func (p *PnpmAdapter) RunAudit(ctx context.Context, projectPath string) ([]models.SecurityIssue, error) {
	cmd := exec.CommandContext(ctx, "pnpm", "audit", "--json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to run pnpm audit: %w", err)
		}
	}

	var auditResult pnpmAuditResult
	if err := json.Unmarshal(output, &auditResult); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm audit output: %w", err)
	}

	var issues []models.SecurityIssue
	for _, vuln := range auditResult.Advisories {
		issue := models.SecurityIssue{
			Severity:    vuln.Severity,
			Title:       vuln.Title,
			Description: vuln.Overview,
			URL:         vuln.URL,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

func (p *PnpmAdapter) GetOutdatedPackages(ctx context.Context, projectPath string) (map[string]*models.OutdatedInfo, error) {
	cmd := exec.CommandContext(ctx, "pnpm", "outdated", "--format", "json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return make(map[string]*models.OutdatedInfo), nil
		}
	}

	var outdatedResult []pnpmOutdatedPackage
	if err := json.Unmarshal(output, &outdatedResult); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm outdated output: %w", err)
	}

	result := make(map[string]*models.OutdatedInfo)
	for _, pkg := range outdatedResult {
		result[pkg.PackageName] = &models.OutdatedInfo{
			Current: pkg.Current,
			Latest:  pkg.Latest,
		}
	}

	return result, nil
}

func (p *PnpmAdapter) readPackageJSON(projectPath string) (*packageJSON, error) {
	pkgPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

func (p *PnpmAdapter) calculatePackageSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

type pnpmListResult struct {
	Name         string                   `json:"name"`
	Version      string                   `json:"version"`
	Dependencies map[string]*pnpmDependency `json:"dependencies"`
}

type pnpmDependency struct {
	Version      string                   `json:"version"`
	Path         string                   `json:"path"`
	Dependencies map[string]*pnpmDependency `json:"dependencies"`
}

type pnpmAuditResult struct {
	Advisories map[string]pnpmAdvisory `json:"advisories"`
}

type pnpmAdvisory struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Overview string `json:"overview"`
	URL      string `json:"url"`
}

type pnpmOutdatedPackage struct {
	PackageName string `json:"packageName"`
	Current     string `json:"current"`
	Latest      string `json:"latest"`
}
