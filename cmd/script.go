package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptTimeout int
)

// scriptCmd represents the script command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Manage and execute scripts",
	Long:  `List, execute, and manage your personal scripts stored in berga.`,
	Aliases: []string{"s"},
}

// scriptListCmd lists available scripts
var scriptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scripts",
	Long:  `Display all available scripts in your berga scripts directory.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return listScripts()
	},
}

// scriptRunCmd runs a script
var scriptRunCmd = &cobra.Command{
	Use:   "run [script-name] [args...]",
	Short: "Execute a script",
	Long:  `Execute a script from your berga scripts directory with optional arguments.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptName := args[0]
		scriptArgs := args[1:]
		return runScript(scriptName, scriptArgs)
	},
}

// scriptEditCmd opens a script for editing
var scriptEditCmd = &cobra.Command{
	Use:   "edit [script-name]",
	Short: "Edit a script",
	Long:  `Open a script for editing using your configured editor.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return editScript(args[0])
	},
}

// scriptShowCmd shows script content
var scriptShowCmd = &cobra.Command{
	Use:   "show [script-name]",
	Short: "Show script content",
	Long:  `Display the content of a script.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showScript(args[0])
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)
	scriptCmd.AddCommand(scriptListCmd)
	scriptCmd.AddCommand(scriptRunCmd)
	scriptCmd.AddCommand(scriptEditCmd)
	scriptCmd.AddCommand(scriptShowCmd)

	// Flags
	scriptRunCmd.Flags().IntVar(&scriptTimeout, "timeout", 300, "Script execution timeout in seconds")
}

func listScripts() error {
	scriptsDir := GetScriptsDir()
	
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		fmt.Printf("Scripts directory does not exist: %s\n", scriptsDir)
		fmt.Println("Run 'berga config init' to initialize your configuration.")
		return nil
	}

	files, err := os.ReadDir(scriptsDir)
	if err != nil {
		return fmt.Errorf("failed to read scripts directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No scripts found.")
		fmt.Printf("Add scripts to: %s\n", scriptsDir)
		return nil
	}

	fmt.Println("Available Scripts:")
	fmt.Println("==================")
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		name := file.Name()
		path := filepath.Join(scriptsDir, name)
		
		// Get file info
		info, err := file.Info()
		if err != nil {
			continue
		}
		
		// Check if executable
		executable := "📄"
		if isExecutable(path) {
			executable = "🚀"
		}
		
		fmt.Printf("  %s %s (%s, %s)\n", 
			executable, 
			name, 
			humanizeSize(info.Size()), 
			info.ModTime().Format("2006-01-02 15:04"))
	}
	
	fmt.Printf("\nScripts directory: %s\n", scriptsDir)
	return nil
}

func runScript(scriptName string, args []string) error {
	scriptsDir := GetScriptsDir()
	scriptPath := filepath.Join(scriptsDir, scriptName)
	
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script '%s' not found in %s", scriptName, scriptsDir)
	}
	
	// Get timeout from config or flag
	timeout := time.Duration(scriptTimeout) * time.Second
	if configTimeout := viper.GetInt("scripts.timeout"); configTimeout > 0 {
		timeout = time.Duration(configTimeout) * time.Second
	}
	
	verbose := viper.GetBool("verbose") || viper.GetBool("scripts.verbose")
	
	if verbose {
		fmt.Printf("Executing: %s %s\n", scriptPath, strings.Join(args, " "))
		fmt.Printf("Timeout: %v\n", timeout)
		fmt.Println("--- Output ---")
	}
	
	// Determine how to execute the script
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, try to execute directly first
		if strings.HasSuffix(strings.ToLower(scriptName), ".ps1") {
			cmd = exec.Command("powershell", append([]string{"-File", scriptPath}, args...)...)
		} else if strings.HasSuffix(strings.ToLower(scriptName), ".bat") || strings.HasSuffix(strings.ToLower(scriptName), ".cmd") {
			cmd = exec.Command("cmd", append([]string{"/C", scriptPath}, args...)...)
		} else {
			// Try to execute directly
			cmd = exec.Command(scriptPath, args...)
		}
	} else {
		// On Unix-like systems, check if it's executable
		if isExecutable(scriptPath) {
			cmd = exec.Command(scriptPath, args...)
		} else {
			// Try to determine the interpreter from the shebang
			interpreter := getInterpreter(scriptPath)
			if interpreter != "" {
				cmd = exec.Command(interpreter, append([]string{scriptPath}, args...)...)
			} else {
				cmd = exec.Command("sh", append([]string{scriptPath}, args...)...)
			}
		}
	}
	
	// Set up the command
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Execute with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()
	
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}
	case <-time.After(timeout):
		cmd.Process.Kill()
		return fmt.Errorf("script execution timed out after %v", timeout)
	}
	
	if verbose {
		fmt.Println("--- Script completed successfully ---")
	}
	
	return nil
}

func editScript(scriptName string) error {
	scriptsDir := GetScriptsDir()
	scriptPath := filepath.Join(scriptsDir, scriptName)
	
	// Get editor from config
	editor := viper.GetString("editor")
	if editor == "" {
		// Try environment variables
		editor = os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			// Default editors by platform
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "nano"
			}
		}
	}
	
	fmt.Printf("Opening %s with %s...\n", scriptPath, editor)
	
	cmd := exec.Command(editor, scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

func showScript(scriptName string) error {
	scriptsDir := GetScriptsDir()
	scriptPath := filepath.Join(scriptsDir, scriptName)
	
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script '%s' not found in %s", scriptName, scriptsDir)
	}
	
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}
	
	fmt.Printf("Script: %s\n", scriptPath)
	fmt.Println("=" + strings.Repeat("=", len(scriptPath)+8))
	fmt.Print(string(content))
	
	return nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	if runtime.GOOS == "windows" {
		// On Windows, check file extension
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".exe" || ext == ".bat" || ext == ".cmd" || ext == ".ps1"
	}
	
	// On Unix-like systems, check execute permission
	return info.Mode()&0111 != 0
}

func getInterpreter(scriptPath string) string {
	file, err := os.Open(scriptPath)
	if err != nil {
		return ""
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#!") {
			shebang := strings.TrimSpace(line[2:])
			// Extract just the interpreter name
			parts := strings.Fields(shebang)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	
	return ""
}

func humanizeSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
}
