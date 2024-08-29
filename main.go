package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	target1    string
	target2    string
	verbose    bool
	outputFile string
	version    = "0.0.32"
)

const (
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
)

func getChangedFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "diff", "--name-only", "@{1}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	changedFiles := strings.Split(strings.TrimSpace(out.String()), "\n")
	return changedFiles, nil
}

func findCommonFiles(changedFiles []string, targetDir string) ([]string, []string, error) {
	var commonFiles []string
	var notFoundFiles []string

	for _, file := range changedFiles {
		fullPath := filepath.Join(targetDir, file)
		if _, err := os.Stat(fullPath); err == nil {
			commonFiles = append(commonFiles, file)
		} else {
			notFoundFiles = append(notFoundFiles, file)
		}
	}

	return commonFiles, notFoundFiles, nil
}

func writeOutputToFile(output string, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(output)
	if err != nil {
		return err
	}
	return nil
}

func runDiffCheck(cmd *cobra.Command, args []string) {
	// Ensure both targets are provided
	if target1 == "" || target2 == "" {
		fmt.Println("Error: both target and source must be specified")
		cmd.Help()
		os.Exit(1)
	}

	// Get the list of changed files from the first target
	changedFiles, err := getChangedFiles(target1)
	if err != nil {
		fmt.Printf("Error getting changed files: %v\n", err)
		os.Exit(1)
	}

	var outputBuilder strings.Builder

	// Verbose output: Print all changed files from target1
	if verbose {
		outputBuilder.WriteString("Changed files from Original Target Project:\n")
		for _, file := range changedFiles {
			outputBuilder.WriteString(fmt.Sprintf("%s\n", file))
		}
	}

	// Find common and not found files in the second target
	commonFiles, notFoundFiles, err := findCommonFiles(changedFiles, target2)
	if err != nil {
		fmt.Printf("Error finding common files: %v\n", err)
		os.Exit(1)
	}

	// Verbose output: Print files not found in target2
	if verbose {
		outputBuilder.WriteString("\nFiles not found in Updated Source Project:\n")
		for _, file := range notFoundFiles {
			outputBuilder.WriteString(fmt.Sprintf(Yellow+"%s"+Reset+"\n", file))
		}

		// Print files found in both targets
		outputBuilder.WriteString("\nFiles found in both projects:\n")
		for _, file := range commonFiles {
			outputBuilder.WriteString(fmt.Sprintf(Red+"%s"+Reset+"\n", file))
		}
	}

	// Print the final list of common files
	if len(commonFiles) > 0 {
		outputBuilder.WriteString("\nFinal list of common files:\n")
		for _, file := range commonFiles {
			outputBuilder.WriteString(fmt.Sprintf("%s\n", file))
		}
	} else {
		outputBuilder.WriteString("No common files found.\n")
	}

	output := outputBuilder.String()

	// If outputFile is specified, write to file
	if outputFile != "" {
		err := writeOutputToFile(output, outputFile)
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", outputFile)
	} else {
		fmt.Print(output)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "git-upgrade-checker",
		Short: "A tool to check common files between git changes and another directory, perfect for upgrading an ongoing live project",
		Run:   runDiffCheck,
	}

	rootCmd.Flags().StringVarP(&target1, "target", "t", "", "Directory of the original target project with git")
	rootCmd.Flags().StringVarP(&target2, "source", "s", "", "Directory of the second project, that has the updated source code")
	rootCmd.Flags().BoolVarP(&verbose, "debug", "d", false, "Enable verbose Debug output")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Specify a file to write the output to")

	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"--version", "-v"},
		Short:   "Print the version number of git-upgrade-checker",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("git-upgrade-checker version", version)
		},
	}

	rootCmd.AddCommand(versionCmd)

	helpCmd := &cobra.Command{
		Use:     "help",
		Aliases: []string{"--help", "-h"},
		Short:   "Show help for git-upgrade-checker",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.Help()
		},
	}

	rootCmd.SetHelpCommand(helpCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

