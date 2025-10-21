package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"aaa-generator/internal/template"
)

func main() {
	var (
		projectName     = flag.String("name", "", "Project name")
		templateName    = flag.String("template", "basic", "Template name")
		listTemplates   = flag.Bool("list", false, "List available templates")
		installTemplate = flag.String("install", "", "Install template from URL or path")
		interactive     = flag.Bool("interactive", false, "Interactive mode")
		version         = flag.Bool("version", false, "Show version")
	)
	flag.Parse()

	if *version {
		fmt.Println("Go React Generator v1.0.0")
		return
	}

	manager, err := template.NewManager()
	if err != nil {
		fmt.Printf("❌ Error initializing template manager: %v\n", err)
		os.Exit(1)
	}

	// 列出可用模板
	if *listTemplates {
		listAvailableTemplates(manager)
		return
	}

	// 安裝模板
	if *installTemplate != "" {
		if err := manager.InstallTemplate(*installTemplate); err != nil {
			fmt.Printf("❌ Error installing template: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Template installed successfully!")
		return
	}

	// 交互模式
	if *interactive {
		if err := runInteractiveMode(manager); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 生成項目
	if *projectName == "" {
		showUsage()
		os.Exit(1)
	}

	generator := template.NewGenerator(manager)
	if err := generator.Generate(*projectName, *templateName); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🎉 Project '%s' created successfully using template '%s'!\n", *projectName, *templateName)
	showNextSteps(*projectName)
}

func listAvailableTemplates(manager *template.Manager) {
	templates := manager.ListTemplates()

	if len(templates) == 0 {
		fmt.Println("No templates available.")
		return
	}

	fmt.Println("📋 Available Templates:")
	fmt.Println()

	for _, tmpl := range templates {
		fmt.Printf("🔹 %s (%s)\n", tmpl.DisplayName, tmpl.Name)
		fmt.Printf("   %s\n", tmpl.Description)
		fmt.Printf("   Version: %s | Source: %s\n", tmpl.Version, tmpl.Source)
		if len(tmpl.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(tmpl.Tags, ", "))
		}
		fmt.Println()
	}
}

func runInteractiveMode(manager *template.Manager) error {
	fmt.Println("🚀 Welcome to Go React Generator!")
	fmt.Println()

	// 選擇模板
	templates := manager.ListTemplates()
	if len(templates) == 0 {
		return fmt.Errorf("no templates available")
	}

	fmt.Println("Available templates:")
	for i, tmpl := range templates {
		fmt.Printf("%d) %s - %s\n", i+1, tmpl.DisplayName, tmpl.Description)
	}

	var choice int
	fmt.Print("\nSelect template (1-", len(templates), "): ")
	if _, err := fmt.Scanf("%d", &choice); err != nil || choice < 1 || choice > len(templates) {
		return fmt.Errorf("invalid template selection")
	}

	selectedTemplate := templates[choice-1]

	// 輸入項目名稱
	var projectName string
	fmt.Print("Enter project name: ")
	if _, err := fmt.Scanf("%s", &projectName); err != nil {
		return fmt.Errorf("invalid project name")
	}

	// 生成項目
	generator := template.NewGenerator(manager)
	if err := generator.Generate(projectName, selectedTemplate.Name); err != nil {
		return err
	}

	fmt.Printf("🎉 Project '%s' created successfully!\n", projectName)
	showNextSteps(projectName)
	return nil
}

func showUsage() {
	fmt.Println("Go React Generator - Create Go + React applications")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  generator -name <project-name> [-template <template-name>]")
	fmt.Println("  generator -list")
	fmt.Println("  generator -interactive")
	fmt.Println("  generator -install <url-or-path>")
	fmt.Println("  generator -version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  generator -name my-app")
	fmt.Println("  generator -name my-app -template advanced")
	fmt.Println("  generator -list")
	fmt.Println("  generator -interactive")
}

func showNextSteps(projectName string) {
	fmt.Println()
	fmt.Println("📚 Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  make install    # Install dependencies")
	fmt.Println("  make dev        # Start development servers")
	fmt.Println("  make build      # Build for production")
	fmt.Println()
}
