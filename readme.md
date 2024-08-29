
# git-upgrade-checker

`git-upgrade-checker` is a command-line tool designed to compare changes in a Git-tracked project directory against another directory, typically used when upgrading ongoing live projects. It helps in identifying files that have changed and exist in both directories, highlighting those that are common and those that are missing from the target directory.

## Installation and Setup

### 1. Build the Tool

Ensure you have Go installed on your machine. Clone the repository containing the `git-upgrade-checker` source code and navigate to the directory. Then, build the tool using the following command:

```bash
go build -o git-upgrade-checker
```

This will generate an executable named `git-upgrade-checker` in your current directory.

### 2. Add Command to PATH

To make the command globally accessible from anywhere in your terminal, move the executable to a directory included in your system's PATH, or add the current directory to the PATH. For example:

```bash
sudo mv git-upgrade-checker /usr/local/bin/
```

Alternatively, you can execute it directly from the current directory using `./git-upgrade-checker`.

## Usage

The `git-upgrade-checker` tool provides various commands and options to customize the behavior. Below are the primary commands and flags:

### Main Command

```bash
git-upgrade-checker [flags]
```

This command compares the changes in one Git-tracked directory (`target`) against another directory (`source`).

**Flags**:

- `-t, --target`: (Required) Path to the original target directory with Git tracking.
- `-s, --source`: (Required) Path to the second directory containing the updated source code.
- `-d, --debug`: (Optional) Enables verbose output, showing detailed information about the files.
- `-o, --output`: (Optional) Specifies a file path to write the output to. If not specified, output is printed to the console.

**Example**:

```bash
git-upgrade-checker -t /path/to/original/project -s /path/to/updated/project
```

This compares the changes in `/path/to/original/project` against `/path/to/updated/project`.

### Version Command

```bash
git-upgrade-checker version
```

Displays the version of the tool.

### Help Command

```bash
git-upgrade-checker help
```

Displays help information about the tool and its usage.

## Adding `git-upgrade-checker` as a Post-Build Command

To automate running `git-upgrade-checker` after building your project, you can add it to your build pipeline or scripts. Here’s how you can integrate it into a Makefile or a simple shell script.

### Example in a Makefile

```make
build:
    go build -o myproject
    git-upgrade-checker -t /path/to/original/project -s /path/to/updated/project
```

### Example in a Shell Script

```bash
#!/bin/bash

# Build your Go project
go build -o myproject

# Run git-upgrade-checker
git-upgrade-checker -t /path/to/original/project -s /path/to/updated/project
```

Make sure to give the script executable permissions:

```bash
chmod +x build-and-check.sh
```

You can now run the script with:

```bash
./build-and-check.sh
```

This will build your project and then run the `git-upgrade-checker` tool as part of your post-build process.

## Conclusion

With the `git-upgrade-checker`, you can easily manage and verify changes across different project directories, ensuring consistency and avoiding potential issues during project upgrades. The tool’s flexibility allows it to be integrated into various workflows, making it a valuable addition to your development toolkit.

