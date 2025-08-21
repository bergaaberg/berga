package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage and use templates",
	Long:  `List, apply, and manage configuration templates stored in berga.`,
	Aliases: []string{"tmpl", "t"},
}

// templateListCmd lists available templates
var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `Display all available templates in your berga templates directory.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return listTemplates()
	},
}

// templateApplyCmd applies a template
var templateApplyCmd = &cobra.Command{
	Use:   "apply [template-name] [output-file]",
	Short: "Apply a template to create a file",
	Long:  `Apply a template with variable substitution to create a new file.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		outputFile := args[1]
		return applyTemplate(templateName, outputFile)
	},
}

// templateShowCmd shows template content
var templateShowCmd = &cobra.Command{
	Use:   "show [template-name]",
	Short: "Show template content",
	Long:  `Display the content of a template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showTemplate(args[0])
	},
}

// templateEditCmd opens a template for editing
var templateEditCmd = &cobra.Command{
	Use:   "edit [template-name]",
	Short: "Edit a template",
	Long:  `Open a template for editing using your configured editor.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return editTemplate(args[0])
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateApplyCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateEditCmd)
}

func listTemplates() error {
	templatesDir := GetTemplatesDir()
	
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		fmt.Printf("Templates directory does not exist: %s\n", templatesDir)
		fmt.Println("Run 'berga config init' to initialize your configuration.")
		return nil
	}

	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No templates found.")
		fmt.Printf("Add templates to: %s\n", templatesDir)
		return nil
	}

	fmt.Println("Available Templates:")
	fmt.Println("===================")
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		name := file.Name()
		
		// Get file info
		info, err := file.Info()
		if err != nil {
			continue
		}
		
		// Remove .tmpl extension for display if present
		displayName := name
		if strings.HasSuffix(name, ".tmpl") {
			displayName = strings.TrimSuffix(name, ".tmpl")
		}
		
		fmt.Printf("  📋 %s (%s, %s)\n", 
			displayName, 
			humanizeSize(info.Size()), 
			info.ModTime().Format("2006-01-02 15:04"))
	}
	
	fmt.Printf("\nTemplates directory: %s\n", templatesDir)
	return nil
}

func applyTemplate(templateName string, outputFile string) error {
	templatesDir := GetTemplatesDir()
	
	// Try to find template file with or without .tmpl extension
	templatePath := filepath.Join(templatesDir, templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join(templatesDir, templateName+".tmpl")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template '%s' not found in %s", templateName, templatesDir)
		}
	}
	
	// Check if output file already exists
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("File %s already exists. Overwrite? (y/N): ", outputFile)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Template application cancelled.")
			return nil
		}
	}
	
	// Read template content
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	
	// Parse template
	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Collect template variables
	vars := collectTemplateVars()
	
	// Create output file
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()
	
	// Execute template
	if err := tmpl.Execute(output, vars); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	
	fmt.Printf("Template '%s' applied successfully to '%s'\n", templateName, outputFile)
	return nil
}

func showTemplate(templateName string) error {
	templatesDir := GetTemplatesDir()
	
	// Try to find template file with or without .tmpl extension
	templatePath := filepath.Join(templatesDir, templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join(templatesDir, templateName+".tmpl")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template '%s' not found in %s", templateName, templatesDir)
		}
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	
	fmt.Printf("Template: %s\n", templatePath)
	fmt.Println("=" + strings.Repeat("=", len(templatePath)+10))
	fmt.Print(string(content))
	
	return nil
}

func editTemplate(templateName string) error {
	templatesDir := GetTemplatesDir()
	
	// Try to find template file with or without .tmpl extension
	templatePath := filepath.Join(templatesDir, templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join(templatesDir, templateName+".tmpl")
	}
	
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
			if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
				editor = "notepad"
			} else {
				editor = "nano"
			}
		}
	}
	
	fmt.Printf("Opening %s with %s...\n", templatePath, editor)
	
	// For new templates, ensure directory exists
	if err := os.MkdirAll(filepath.Dir(templatePath), 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}
	
	cmd := exec.Command(editor, templatePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

func collectTemplateVars() map[string]interface{} {
	vars := make(map[string]interface{})
	
	// Get common variables from config
	vars["Author"] = viper.GetString("templates.author")
	vars["Email"] = viper.GetString("templates.email")
	
	// Add some default variables
	if cwd, err := os.Getwd(); err == nil {
		vars["CurrentDir"] = filepath.Base(cwd)
		vars["ProjectName"] = filepath.Base(cwd)
	}
	
	// Interactive variable collection
	fmt.Println("Template Variables:")
	fmt.Println("==================")
	
	// Prompt for project name if not set
	if vars["ProjectName"] == "" || vars["ProjectName"] == "." {
		fmt.Print("Project Name: ")
		var projectName string
		fmt.Scanln(&projectName)
		if projectName != "" {
			vars["ProjectName"] = projectName
		}
	} else {
		fmt.Printf("Project Name: %s\n", vars["ProjectName"])
	}
	
	// Prompt for author if not set
	if vars["Author"] == "" {
		fmt.Print("Author: ")
		var author string
		fmt.Scanln(&author)
		if author != "" {
			vars["Author"] = author
		}
	} else {
		fmt.Printf("Author: %s\n", vars["Author"])
	}
	
	// Prompt for additional custom variables
	fmt.Print("Additional variables (key=value, empty to finish): ")
	for {
		var input string
		fmt.Scanln(&input)
		if input == "" {
			break
		}
		
		parts := strings.SplitN(input, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
		
		fmt.Print("Additional variables (key=value, empty to finish): ")
	}
	
	return vars
}
