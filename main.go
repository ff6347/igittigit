package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "--help", "-h", "help":
		printUsage()
	case "list":
		handleList()
	case "search":
		handleSearch()
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("Error: template name required")
			fmt.Println("Usage: igittigit get <template>")
			os.Exit(1)
		}
		template := os.Args[2]
		handleGet(template)
	case "completions":
		handleCompletions()
	default:
		if err := runPath(os.Args[1:], os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("igittigit - Literal ignore rules and GitHub .gitignore templates")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  igittigit [--here] [--allow] [--] <path>")
	fmt.Println("                             Append one literal path rule (requires Git)")
	fmt.Println("  igittigit list              List all available templates")
	fmt.Println("  igittigit search            Interactive fuzzy search (requires fzf)")
	fmt.Println("  igittigit get <template>    Fetch specific template")
	fmt.Println("  igittigit completions       Generate fish shell completions")
	fmt.Println("")
	fmt.Println("Path rules:")
	fmt.Println("  Default: nearest ancestor .gitignore in the target's working tree.")
	fmt.Println("  If absent, offers to create one at the working-tree root (default No).")
	fmt.Println("  --here: use the shell working directory's .gitignore; target must be below it.")
	fmt.Println("  --allow: append an exception; ignored parents can still block it (exit 1).")
	fmt.Println("  Tracked files remain tracked; directory exceptions do not allow every descendant.")
	fmt.Println("  Missing paths are allowed; a trailing / marks a future directory.")
	fmt.Println("  Target links mean the link itself; linked-directory traversal is rejected.")
	fmt.Println("  Symbolic-link ignore files are refused. Git administrative paths are rejected.")
	fmt.Println("  --here outside Git takes effect once the directory belongs to a working tree.")
	fmt.Println("  Otherwise outside Git: offers an escaped basename global pattern, affecting")
	fmt.Println("  every repository, at Git's configured/default excludes file (default No).")
	fmt.Println("  Global exceptions cannot override local exclusions. Flags precede the path.")
	fmt.Println("  Use ./list for a path named list, or -- -name for a leading dash.")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  igittigit ./build/")
	fmt.Println("  igittigit --here ./build/")
	fmt.Println("  igittigit --allow ./config/example.env")
	fmt.Println("  igittigit --here --allow ./config/example.env")
	fmt.Println("  igittigit list")
	fmt.Println("  igittigit search")
	fmt.Println("  igittigit get Go")
	fmt.Println("  igittigit get Node > .gitignore")
}

func handleList() {
	templates, err := getTemplateList()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	for _, template := range templates {
		fmt.Println(template)
	}
}

func handleSearch() {
	if !hasFzf() {
		fmt.Println("Error: fzf not found. Please install fzf for interactive search.")
		fmt.Println("Visit: https://github.com/junegunn/fzf#installation")
		os.Exit(1)
	}

	templates, err := getTemplateList()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	selected, err := runFzf(templates)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if selected == "" {
		os.Exit(0)
	}

	handleGetInteractive(selected)
}

func handleGet(template string) {
	content, err := getTemplate(template)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(content)
}

func handleGetInteractive(template string) {
	if fileExists(".gitignore") {
		action := promptForAction()
		switch action {
		case "cancel":
			fmt.Println("Cancelled.")
			os.Exit(0)
		case "overwrite":
			writeGitignore(template, false)
		case "append":
			writeGitignore(template, true)
		}
	} else {
		writeGitignore(template, false)
	}
}

func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func runFzf(items []string) (string, error) {
	cmd := exec.Command("fzf", "--height", "40%", "--reverse", "--border")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		for _, item := range items {
			fmt.Fprintln(stdin, item)
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 130 {
				return "", nil
			}
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func promptForAction() string {
	fmt.Println(".gitignore already exists. What would you like to do?")
	fmt.Println("  [o] Overwrite")
	fmt.Println("  [a] Append")
	fmt.Println("  [c] Cancel")
	fmt.Print("Choice (o/a/c): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "cancel"
	}

	choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
	switch choice {
	case "o", "overwrite":
		return "overwrite"
	case "a", "append":
		return "append"
	case "c", "cancel", "":
		return "cancel"
	default:
		fmt.Printf("Invalid choice '%s'. Cancelled.\n", choice)
		return "cancel"
	}
}

func writeGitignore(template string, append bool) {
	content, err := getTemplate(template)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var file *os.File
	if append {
		file, err = os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening .gitignore: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(file, "\n# %s\n%s\n", template, content)
	} else {
		file, err = os.Create(".gitignore")
		if err != nil {
			fmt.Printf("Error creating .gitignore: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(file, "%s\n", content)
	}

	file.Close()

	action := "Created"
	if append {
		action = "Appended to"
	}
	fmt.Printf("%s .gitignore with %s template\n", action, template)
}

func handleCompletions() {
	generateFishCompletions()
}

func generateFishCompletions() {
	fmt.Print(`# Fish completions for igittigit
# Install: igittigit completions > ~/.config/fish/completions/igittigit.fish

function __igittigit_needs_path
    set -l args (commandline -opc)
    set -e args[1]
    set -l flags 1
    for arg in $args
        if test $flags -eq 1
            switch $arg
                case --here --allow
                    continue
                case '--'
                    set flags 0
                    continue
            end
        end
        return 1
    end
    return 0
end

# Complete subcommands
complete -c igittigit -f -n "test (count (commandline -opc)) -eq 1" -a "list" -d "List all available templates"
complete -c igittigit -f -n "test (count (commandline -opc)) -eq 1" -a "search" -d "Interactive fuzzy search (requires fzf)"
complete -c igittigit -f -n "test (count (commandline -opc)) -eq 1" -a "get" -d "Fetch specific template"
complete -c igittigit -f -n "test (count (commandline -opc)) -eq 1" -a "completions" -d "Generate fish shell completions"
complete -c igittigit -f -n "test (count (commandline -opc)) -eq 1" -a "help" -d "Show help information"

# Complete help flags
complete -c igittigit -f -n "test (count (commandline -opc)) -eq 1" -s h -l help -d "Show help information"

complete -c igittigit -F -n "__igittigit_needs_path"
complete -c igittigit -n "__igittigit_needs_path; and not contains -- -- (commandline -opc)" -l here -d "Use the working directory's .gitignore"
complete -c igittigit -n "__igittigit_needs_path; and not contains -- -- (commandline -opc)" -l allow -d "Append an exception rule"

# Complete template names for 'get' command
complete -c igittigit -f -n "__fish_seen_subcommand_from get" -a "(igittigit list 2>/dev/null)"
`)
}
