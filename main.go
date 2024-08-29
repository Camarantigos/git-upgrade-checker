package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

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
	cmd := exec.Command("git", "-C", dir, "diff", "--name-only", "HEAD@{1}")
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

func getFileDiff(dir, file string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "diff", "HEAD@{1}", "--", file)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func formatDiffWithLineNumbers(diff string) string {
	lines := strings.Split(diff, "\n")
	var formattedDiff strings.Builder

	oldLineNum := 0
	newLineNum := 0

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "@@"):
			// Extract line numbers from the hunk header
			var oldStart, newStart int
			fmt.Sscanf(line, "@@ -%d,%*d +%d,%*d @@", &oldStart, &newStart)
			oldLineNum = oldStart - 1
			newLineNum = newStart - 1
			formattedDiff.WriteString(line + "\n")
		case strings.HasPrefix(line, "-"):
			oldLineNum++
			formattedDiff.WriteString(fmt.Sprintf("<<<<<<OLD (line %d)>>>>>> %s\n", oldLineNum, line[1:]))
		case strings.HasPrefix(line, "+"):
			newLineNum++
			formattedDiff.WriteString(fmt.Sprintf("<<<<<<CHANGE (line %d)>>>>>> %s\n", newLineNum, line[1:]))
		default:
			if len(line) > 0 && line[0] != '@' && line[0] != ' ' && line[0] != '\\' {
				oldLineNum++
				newLineNum++
				formattedDiff.WriteString(line + "\n")
			}
		}
	}

	return formattedDiff.String()
}

func calculateColumnWidths(files []string) (int, int, int) {
	maxNumberWidth := len("Number")
	maxFilePathWidth := len("File Path")
	maxFileNameWidth := len("File Name")

	for i, file := range files {
		numberWidth := len(fmt.Sprintf("%d", i+1))
		filePathWidth := len(filepath.Dir(file) + "/")
		fileNameWidth := len(filepath.Base(file))

		if numberWidth > maxNumberWidth {
			maxNumberWidth = numberWidth
		}
		if filePathWidth > maxFilePathWidth {
			maxFilePathWidth = filePathWidth
		}
		if fileNameWidth > maxFileNameWidth {
			maxFileNameWidth = fileNameWidth
		}
	}

	return maxNumberWidth, maxFilePathWidth, maxFileNameWidth
}

func printTable(writer *tabwriter.Writer, files []string, dir, color string) {
	maxNumberWidth, maxFilePathWidth, maxFileNameWidth := calculateColumnWidths(files)

	fmt.Println("┌" + strings.Repeat("─", maxNumberWidth+2) + "┬" + strings.Repeat("─", maxFilePathWidth+2) + "┬" + strings.Repeat("─", maxFileNameWidth+2) + "┬─────────┐")
	fmt.Fprintf(writer, "│ %-*s │ %-*s │ %-*s │ %-7s │\n", maxNumberWidth, "Number", maxFilePathWidth, "File Path", maxFileNameWidth, "File Name", "Changes")
	fmt.Println("├" + strings.Repeat("─", maxNumberWidth+2) + "┼" + strings.Repeat("─", maxFilePathWidth+2) + "┼" + strings.Repeat("─", maxFileNameWidth+2) + "┼─────────┤")

	for i, file := range files {
		filePath := filepath.Dir(file)
		fileName := filepath.Base(file)
		diff, _ := getFileDiff(dir, file)
		formattedDiff := formatDiffWithLineNumbers(diff)
		fmt.Fprintf(writer, color+"│ %-*d │ %-*s │ %-*s │ %-7d │\n"+Reset, maxNumberWidth, i+1, maxFilePathWidth, filePath+"/", maxFileNameWidth, fileName, len(diff))
		fmt.Println(formattedDiff)
	}

	fmt.Println("└" + strings.Repeat("─", maxNumberWidth+2) + "┴" + strings.Repeat("─", maxFilePathWidth+2) + "┴" + strings.Repeat("─", maxFileNameWidth+2) + "┴─────────┘")

	writer.Flush()
}

func printVerboseTableOutput(commonFiles, notFoundFiles []string, dir string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print the common files table
	if len(commonFiles) > 0 {
		fmt.Println("\nFiles found in both projects:")
		printTable(writer, commonFiles, dir, Red)
	}

	// Print the not found files table
	if len(notFoundFiles) > 0 {
		fmt.Println("\nFiles not found in Updated Source Project:")
		printTable(writer, notFoundFiles, dir, Yellow)
	}
}

func printTableOutput(commonFiles []string, dir string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printTable(writer, commonFiles, dir, Reset)
}

func runDiffCheck(cmd *cobra.Command, args []string) {
	if target1 == "" || target2 == "" {
		fmt.Println("Error: both target and source must be specified")
		cmd.Help()
		os.Exit(1)
	}

	changedFiles, err := getChangedFiles(target1)
	if err != nil {
		fmt.Printf("Error getting changed files: %v\n", err)
		os.Exit(1)
	}

	commonFiles, notFoundFiles, err := findCommonFiles(changedFiles, target2)
	if err != nil {
		fmt.Printf("Error finding common files: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		printVerboseTableOutput(commonFiles, notFoundFiles, target1)
	} else if outputFile != "" {
		err := writeCSVOutput(commonFiles, outputFile)
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			os.Exit(1)
		}
	} else {
		printTableOutput(commonFiles, target1)
	}
}

func writeCSVOutput(commonFiles []string, outputFilePath string) error {
	file, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"number", "file_path", "file_name", "changes"})
	if err != nil {
		return err
	}

	for i, file := range commonFiles {
		filePath := filepath.Dir(file)
		fileName := filepath.Base(file)
		diff, _ := getFileDiff(target1, file)
		formattedDiff := formatDiffWithLineNumbers(diff)
		err = writer.Write([]string{
			fmt.Sprintf("%d", i+1),
			filePath + "/",
			fileName,
			formattedDiff,
		})
		if err != nil {
			return err
		}
	}

	return nil
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
