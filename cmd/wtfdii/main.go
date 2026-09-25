package main

import (
	"context"
	"fmt"
	"os"

	"github.com/skytr1x/wtfdii/internal/analyzer"
	"github.com/skytr1x/wtfdii/internal/config"
	"github.com/skytr1x/wtfdii/internal/output"

	"github.com/spf13/cobra"
)

var (
	projectPath string
	noColor     bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "wtfdii",
		Short: "WTF Did I Install? - CLI tool to analyze project dependencies",
		Long: `WTF Did I Install? helps you understand what's in your project's dependencies.
It shows you what packages are installed, why they're there, and potential issues.`,
		RunE: runOverview,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&projectPath, "path", "p", ".", "Project path to analyze")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Subcommands
	rootCmd.AddCommand(whyCmd())
	rootCmd.AddCommand(sizeCmd())
	rootCmd.AddCommand(securityCmd())
	rootCmd.AddCommand(staleCmd())
	rootCmd.AddCommand(unusedCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runOverview(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := output.New(!noColor)

	cfg, err := config.Load(projectPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatter.FormatError(err))
		return nil
	}

	a, err := analyzer.New(projectPath, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatter.FormatError(err))
		return nil
	}

	tree, err := a.Analyze(ctx, projectPath)
	if err != nil {
		cmd.Help()
		fmt.Fprintln(os.Stderr, "\nIf you're running it in a project directory, make sure that there are any packages.")
		return nil
	}

	fmt.Print(formatter.FormatOverview(tree))
	return nil
}

func whyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "why <package>",
		Short: "Show why a package is installed",
		Long:  "Display the dependency chain showing why a specific package is installed in your project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]
			ctx := context.Background()
			formatter := output.New(!noColor)

			cfg, err := config.Load(projectPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			a, err := analyzer.New(projectPath, cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			chain, err := a.GetDependencyChain(ctx, projectPath, packageName)
			if err != nil {
				cmd.Help()
				fmt.Fprintln(os.Stderr, "\nIf you're running it in a project directory, make sure that there are any packages.")
				return nil
			}

			fmt.Print(formatter.FormatDependencyChain(packageName, chain))
			return nil
		},
	}
}

func sizeCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "size",
		Short: "Show largest packages",
		Long:  "Display packages sorted by size, showing which ones take up the most disk space.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := output.New(!noColor)

			cfg, err := config.Load(projectPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			a, err := analyzer.New(projectPath, cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			packages, err := a.GetPackagesBySize(ctx, projectPath, limit)
			if err != nil {
				cmd.Help()
				fmt.Fprintln(os.Stderr, "\nIf you're running it in a project directory, make sure that there are any packages.")
				return nil
			}

			fmt.Print(formatter.FormatPackagesBySize(packages, limit))
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Number of packages to show")
	return cmd
}

func securityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "security",
		Short: "Run security audit",
		Long:  "Analyze dependencies for known security vulnerabilities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := output.New(!noColor)

			cfg, err := config.Load(projectPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			a, err := analyzer.New(projectPath, cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			issues, err := a.GetSecurityIssues(ctx, projectPath)
			if err != nil {
				cmd.Help()
				fmt.Fprintln(os.Stderr, "\nIf you're running it in a project directory, make sure that there are any packages.")
				return nil
			}

			fmt.Print(formatter.FormatSecurityIssues(issues))
			return nil
		},
	}
}

func staleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stale",
		Short: "Show outdated packages",
		Long:  "Display packages that have newer versions available.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := output.New(!noColor)

			cfg, err := config.Load(projectPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			a, err := analyzer.New(projectPath, cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			packages, err := a.GetStalePackages(ctx, projectPath)
			if err != nil {
				cmd.Help()
				fmt.Fprintln(os.Stderr, "\nIf you're running it in a project directory, make sure that there are any packages.")
				return nil
			}

			fmt.Print(formatter.FormatStalePackages(packages))
			return nil
		},
	}
}

func unusedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unused",
		Short: "Show potentially unused dependencies",
		Long:  "Analyze source code to detect dependencies that may not be used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := output.New(!noColor)

			cfg, err := config.Load(projectPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			a, err := analyzer.New(projectPath, cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, formatter.FormatError(err))
				return nil
			}

			packages, err := a.GetUnusedPackages(ctx, projectPath)
			if err != nil {
				cmd.Help()
				fmt.Fprintln(os.Stderr, "\nIf you're running it in a project directory, make sure that there are any packages.")
				return nil
			}

			fmt.Print(formatter.FormatUnusedPackages(packages))
			return nil
		},
	}
}
