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

type YarnAdapter struct{}

func (y *YarnAdapter) Name() string {
	return "yarn"
}

func (y *YarnAdapter) Detect(projectPath string) (bool, error) {
	lockPath := filepath.Join(projectPath, "yarn.lock")
	if _, err := os.Stat(lockPath); err == nil {
		return true, nil
	}
	return false, nil
}

func (y *YarnAdapter) GetDependencies(ctx context.Context, projectPath string) (*models.DependencyTree, error) {
	pkgPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package.json not found in %s\n\nMake sure you're running wtfdii in a Node.js project directory", projectPath)
	}

	tree := &models.DependencyTree{
		PackageManager: "yarn",
		Packages:       make(map[string]*models.Package),
	}

	packageJSON, err := y.readPackageJSON(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	tree.ProjectName = packageJSON.Name
	cmd := exec.CommandContext(ctx, "yarn", "list", "--json", "--depth", "999")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to get yarn list: %w", err)
		}
	}

	lines := strings.Split(string(output), "\n")
	directDeps := make(map[string]bool)
	for name := range packageJSON.Dependencies {
		directDeps[name] = true
	}
	for name := range packageJSON.DevDependencies {
		directDeps[name] = true
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry yarnListEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Type == "tree" && entry.Data.Name != "" {
			parts := strings.Split(entry.Data.Name, "@")
			var name, version string
			if strings.HasPrefix(entry.Data.Name, "@") {
				if len(parts) >= 3 {
					name = "@" + parts[1]
					version = parts[2]
				}
			} else {
				if len(parts) >= 2 {
					name = parts[0]
					version = parts[1]
				}
			}

			if name != "" {
				pkg := &models.Package{
					Name:         name,
					Version:      version,
					IsDirect:     directDeps[name],
					IsTransitive: !directDeps[name],
					DependsOn:    []string{},
					DependedOnBy: []string{},
				}

				nodeModulesPath := filepath.Join(projectPath, "node_modules", name)
				if size, err := y.calculatePackageSize(nodeModulesPath); err == nil {
					pkg.Size = size
				}

				tree.Packages[name] = pkg
			}
		}
	}

	tree.DirectCount = len(directDeps)
	tree.TransitiveCount = len(tree.Packages) - tree.DirectCount
	tree.TotalCount = len(tree.Packages)

	for _, pkg := range tree.Packages {
		tree.TotalSize += pkg.Size
	}

	return tree, nil
}

func (y *YarnAdapter) GetPackageInfo(ctx context.Context, projectPath string, packageName string) (*models.Package, error) {
	tree, err := y.GetDependencies(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	pkg, exists := tree.Packages[packageName]
	if !exists {
		return nil, fmt.Errorf("package %s not found", packageName)
	}

	return pkg, nil
}

func (y *YarnAdapter) GetDependencyChain(ctx context.Context, projectPath string, packageName string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "yarn", "why", packageName)
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency chain: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var chain []string

	for _, line := range lines {
		if strings.Contains(line, "=>") || strings.Contains(line, packageName) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				chain = append(chain, strings.Trim(parts[0], `"'`))
			}
		}
	}

	return chain, nil
}

func (y *YarnAdapter) GetPackageSize(projectPath string, packageName string) (int64, error) {
	nodeModulesPath := filepath.Join(projectPath, "node_modules", packageName)
	return y.calculatePackageSize(nodeModulesPath)
}

func (y *YarnAdapter) RunAudit(ctx context.Context, projectPath string) ([]models.SecurityIssue, error) {
	cmd := exec.CommandContext(ctx, "yarn", "audit", "--json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to run yarn audit: %w", err)
		}
	}

	var issues []models.SecurityIssue
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry yarnAuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Type == "auditAdvisory" {
			issue := models.SecurityIssue{
				Severity:    entry.Data.Advisory.Severity,
				Title:       entry.Data.Advisory.Title,
				Description: entry.Data.Advisory.Overview,
				URL:         entry.Data.Advisory.URL,
			}
			issues = append(issues, issue)
		}
	}

	return issues, nil
}

func (y *YarnAdapter) GetOutdatedPackages(ctx context.Context, projectPath string) (map[string]*models.OutdatedInfo, error) {
	cmd := exec.CommandContext(ctx, "yarn", "outdated", "--json")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		if len(output) == 0 {
			return make(map[string]*models.OutdatedInfo), nil
		}
	}

	result := make(map[string]*models.OutdatedInfo)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry yarnOutdatedEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Type == "table" && entry.Data.Head != nil {
			for _, row := range entry.Data.Body {
				if len(row) >= 3 {
					result[row[0]] = &models.OutdatedInfo{
						Current: row[1],
						Latest:  row[3],
					}
				}
			}
		}
	}

	return result, nil
}

func (y *YarnAdapter) readPackageJSON(projectPath string) (*packageJSON, error) {
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

func (y *YarnAdapter) calculatePackageSize(path string) (int64, error) {
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

type yarnListEntry struct {
	Type string `json:"type"`
	Data struct {
		Name string `json:"name"`
	} `json:"data"`
}

type yarnAuditEntry struct {
	Type string `json:"type"`
	Data struct {
		Advisory struct {
			Severity string `json:"severity"`
			Title    string `json:"title"`
			Overview string `json:"overview"`
			URL      string `json:"url"`
		} `json:"advisory"`
	} `json:"data"`
}

type yarnOutdatedEntry struct {
	Type string `json:"type"`
	Data struct {
		Head []string   `json:"head"`
		Body [][]string `json:"body"`
	} `json:"data"`
}
