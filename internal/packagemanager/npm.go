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

type NpmAdapter struct{}

func (n *NpmAdapter) Name() string {
	return "npm"
}

func (n *NpmAdapter) Detect(projectPath string) (bool, error) {
	lockPath := filepath.Join(projectPath, "package-lock.json")
	if _, err := os.Stat(lockPath); err == nil {
		return true, nil
	}

	pkgPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		pnpmLock := filepath.Join(projectPath, "pnpm-lock.yaml")
		yarnLock := filepath.Join(projectPath, "yarn.lock")

		if _, err := os.Stat(pnpmLock); err != nil {
			if _, err := os.Stat(yarnLock); err != nil {
				return true, nil
			}
		}
	}

	return false, nil
}

func (n *NpmAdapter) GetDependencies(ctx context.Context, projectPath string) (*models.DependencyTree, error) {
	pkgPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package.json not found in %s\n\nMake sure you're running wtfdii in a Node.js project directory", projectPath)
	}

	tree := &models.DependencyTree{
		PackageManager: "npm",
		Packages:       make(map[string]*models.Package),
	}

	packageJSON, err := n.readPackageJSON(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	tree.ProjectName = packageJSON.Name
	cmd := exec.CommandContext(ctx, "npm", "list", "--json", "--all", "--long")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to get npm list: %w", err)
		}
	}

	var listResult npmListResult
	if err := json.Unmarshal(output, &listResult); err != nil {
		return nil, fmt.Errorf("failed to parse npm list output: %w", err)
	}

	directDeps := make(map[string]bool)
	for name := range packageJSON.Dependencies {
		directDeps[name] = true
	}
	for name := range packageJSON.DevDependencies {
		directDeps[name] = true
	}

	n.parseDependencies(&listResult, tree, directDeps, "")
	tree.DirectCount = len(directDeps)
	tree.TransitiveCount = len(tree.Packages) - tree.DirectCount
	tree.TotalCount = len(tree.Packages)
	for _, pkg := range tree.Packages {
		tree.TotalSize += pkg.Size
	}

	return tree, nil
}

func (n *NpmAdapter) parseDependencies(node *npmListResult, tree *models.DependencyTree, directDeps map[string]bool, parent string) {
	for name, dep := range node.Dependencies {
		pkg := &models.Package{
			Name:         name,
			Version:      dep.Version,
			IsDirect:     directDeps[name],
			IsTransitive: !directDeps[name],
			DependsOn:    []string{},
			DependedOnBy: []string{},
		}

		if parent != "" {
			pkg.DependedOnBy = append(pkg.DependedOnBy, parent)
		}

		if dep.Path != "" {
			if size, err := n.calculatePackageSize(dep.Path); err == nil {
				pkg.Size = size
			}
		}

		tree.Packages[name] = pkg

		if len(dep.Dependencies) > 0 {
			n.parseDependencies(&npmListResult{Dependencies: dep.Dependencies}, tree, directDeps, name)
		}
	}
}

func (n *NpmAdapter) GetPackageInfo(ctx context.Context, projectPath string, packageName string) (*models.Package, error) {
	tree, err := n.GetDependencies(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	pkg, exists := tree.Packages[packageName]
	if !exists {
		return nil, fmt.Errorf("package %s not found", packageName)
	}

	return pkg, nil
}

func (n *NpmAdapter) GetDependencyChain(ctx context.Context, projectPath string, packageName string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "npm", "ls", packageName, "--all")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency chain: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var chain []string

	for _, line := range lines {
		if strings.Contains(line, packageName) {
			// Extract package names from the tree structure
			parts := strings.Split(line, "─")
			if len(parts) > 0 {
				pkgInfo := strings.TrimSpace(parts[len(parts)-1])
				pkgName := strings.Split(pkgInfo, "@")[0]
				chain = append(chain, pkgName)
			}
		}
	}

	return chain, nil
}

func (n *NpmAdapter) GetPackageSize(projectPath string, packageName string) (int64, error) {
	nodeModulesPath := filepath.Join(projectPath, "node_modules", packageName)
	return n.calculatePackageSize(nodeModulesPath)
}

func (n *NpmAdapter) RunAudit(ctx context.Context, projectPath string) ([]models.SecurityIssue, error) {
	cmd := exec.CommandContext(ctx, "npm", "audit", "--json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to run npm audit: %w", err)
		}
	}

	var auditResult npmAuditResult
	if err := json.Unmarshal(output, &auditResult); err != nil {
		return nil, fmt.Errorf("failed to parse npm audit output: %w", err)
	}

	var issues []models.SecurityIssue
	for _, vuln := range auditResult.Vulnerabilities {
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

func (n *NpmAdapter) GetOutdatedPackages(ctx context.Context, projectPath string) (map[string]*models.OutdatedInfo, error) {
	cmd := exec.CommandContext(ctx, "npm", "outdated", "--json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return make(map[string]*models.OutdatedInfo), nil
		}
	}

	var outdatedResult map[string]npmOutdatedPackage
	if err := json.Unmarshal(output, &outdatedResult); err != nil {
		return nil, fmt.Errorf("failed to parse npm outdated output: %w", err)
	}

	result := make(map[string]*models.OutdatedInfo)
	for name, pkg := range outdatedResult {
		result[name] = &models.OutdatedInfo{
			Current:  pkg.Current,
			Latest:   pkg.Latest,
			Type:     pkg.Type,
			Homepage: pkg.Homepage,
		}
	}

	return result, nil
}

func (n *NpmAdapter) readPackageJSON(projectPath string) (*packageJSON, error) {
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

func (n *NpmAdapter) calculatePackageSize(path string) (int64, error) {
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

type packageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type npmListResult struct {
	Name         string                   `json:"name"`
	Version      string                   `json:"version"`
	Dependencies map[string]*npmDependency `json:"dependencies"`
}

type npmDependency struct {
	Version      string                   `json:"version"`
	Path         string                   `json:"path"`
	Dependencies map[string]*npmDependency `json:"dependencies"`
}

type npmAuditResult struct {
	Vulnerabilities map[string]npmVulnerability `json:"vulnerabilities"`
}

type npmVulnerability struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Overview string `json:"overview"`
	URL      string `json:"url"`
}

type npmOutdatedPackage struct {
	Current  string `json:"current"`
	Latest   string `json:"latest"`
	Type     string `json:"type"`
	Homepage string `json:"homepage"`
}
