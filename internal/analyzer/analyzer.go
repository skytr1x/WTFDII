package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/skytr1x/wtfdii/internal/models"
	"github.com/skytr1x/wtfdii/internal/packagemanager"
)

type Analyzer struct {
	manager packagemanager.PackageManager
	config  *models.Config
}

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

func (a *Analyzer) GetDependencyChain(ctx context.Context, projectPath string, packageName string) ([]string, error) {
	return a.manager.GetDependencyChain(ctx, projectPath, packageName)
}

func (a *Analyzer) GetPackagesBySize(ctx context.Context, projectPath string, limit int) ([]*models.Package, error) {
	tree, err := a.Analyze(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	packages := make([]*models.Package, 0, len(tree.Packages))
	for _, pkg := range tree.Packages {
		packages = append(packages, pkg)
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Size > packages[j].Size
	})

	if limit > 0 && limit < len(packages) {
		packages = packages[:limit]
	}

	return packages, nil
}

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

func (a *Analyzer) GetSecurityIssues(ctx context.Context, projectPath string) ([]models.SecurityIssue, error) {
	return a.manager.RunAudit(ctx, projectPath)
}

func (a *Analyzer) GetUnusedPackages(ctx context.Context, projectPath string) ([]*models.Package, error) {
	tree, err := a.Analyze(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	usedPackages, err := a.findUsedPackages(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find used packages: %w", err)
	}

	var unusedPackages []*models.Package
	for name, pkg := range tree.Packages {
		if !pkg.IsDirect {
			continue
		}

		if a.shouldIgnorePackage(name) {
			continue
		}

		if !usedPackages[name] {
			unusedPackages = append(unusedPackages, pkg)
		}
	}

	sort.Slice(unusedPackages, func(i, j int) bool {
		return unusedPackages[i].Name < unusedPackages[j].Name
	})

	return unusedPackages, nil
}

func (a *Analyzer) findUsedPackages(projectPath string) (map[string]bool, error) {
	used := make(map[string]bool)

	extensions := []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"}
	for _, ext := range extensions {
		if err := a.scanFilesForImports(projectPath, ext, used); err != nil {
			return nil, err
		}
	}

	if err := a.checkPackageJSONScripts(projectPath, used); err != nil {
		return nil, err
	}

	return used, nil
}

func (a *Analyzer) scanFilesForImports(projectPath, extension string, used map[string]bool) error {
	cmd := fmt.Sprintf("cd %s && find . -name '*%s' -type f ! -path '*/node_modules/*' ! -path '*/.git/*' -print0 | xargs -0 grep -hoE \"(require\\s*\\(\\s*['\\\"]([^'\\\"]+)['\\\"]|import\\s+.*\\s+from\\s+['\\\"]([^'\\\"]+)['\\\"]|import\\s*\\(\\s*['\\\"]([^'\\\"]+)['\\\"])\" 2>/dev/null || true", projectPath, extension)

	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		pkgName := a.extractPackageName(line)
		if pkgName != "" {
			used[pkgName] = true
		}
	}

	return nil
}

func (a *Analyzer) grepPattern(projectPath, pattern, extension string) ([]string, error) {
	cmd := fmt.Sprintf("cd %s && find . -name '*%s' -type f ! -path '*/node_modules/*' ! -path '*/.git/*' -exec grep -hoE '%s' {} \\; 2>/dev/null || true", projectPath, extension, pattern)

	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results []string
	for _, line := range lines {
		if line != "" {
			results = append(results, line)
		}
	}

	return results, nil
}

func (a *Analyzer) extractPackageName(importLine string) string {
	re := regexp.MustCompile(`['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(importLine)
	if len(matches) < 2 {
		return ""
	}

	pkgPath := matches[1]

	if strings.HasPrefix(pkgPath, ".") || strings.HasPrefix(pkgPath, "/") {
		return ""
	}

	parts := strings.Split(pkgPath, "/")
	if strings.HasPrefix(pkgPath, "@") && len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}

	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

func (a *Analyzer) checkPackageJSONScripts(projectPath string, used map[string]bool) error {
	pkgPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	for _, script := range pkg.Scripts {
		words := strings.Fields(script)
		for _, word := range words {
			word = strings.Trim(word, "\"'")
			if !strings.Contains(word, "/") && !strings.HasPrefix(word, "-") {
				used[word] = true
			}
		}
	}

	return nil
}

func (a *Analyzer) PackageManagerName() string {
	return a.manager.Name()
}
