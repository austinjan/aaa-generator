package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"aaa-generator/internal/template"
	"github.com/spf13/cobra"
)

var version = "dev" // Set via ldflags during build

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printWelcomeBanner() {
	fmt.Println("╭─────────────────────────────────────────────────────────╮")
	fmt.Printf("│              AAA-Generator v%-8s                    │\n", version)
	fmt.Println("│            Application Template Generator               │")
	fmt.Println("╰─────────────────────────────────────────────────────────╯")
	fmt.Println()
}

func newRootCommand() *cobra.Command {
	var (
		projectName   string
		templateName  string
		listFlag      bool
		installTarget string
		interactive   bool
		versionFlag   bool
	)

	cmd := &cobra.Command{
		Use:          "generator",
		Short:        "Create Go + React applications from templates",
		Long:         "Go React Generator scaffolds Go backends and React frontends using reusable templates.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				printWelcomeBanner()
				return nil
			}

			manager, err := template.NewManager()
			if err != nil {
				return fmt.Errorf("error initializing template manager: %w", err)
			}

			if listFlag {
				listAvailableTemplates(manager)
				return nil
			}

			if installTarget != "" {
				fmt.Println()
				fmt.Printf("📦 Installing template from: %s\n", installTarget)
				fmt.Println("───────────────────────────────────────────────────────")
				if err := manager.InstallTemplate(installTarget); err != nil {
					return fmt.Errorf("error installing template: %w", err)
				}
				fmt.Println("✅ Template installed successfully!")
				fmt.Println()
				return nil
			}

			if interactive {
				if err := checkEnvironment(cmd.OutOrStdout()); err != nil {
					return err
				}
				if err := runInteractiveMode(manager); err != nil {
					return err
				}
				return nil
			}

			if projectName == "" {
				return fmt.Errorf("project name is required (use --name or run with --interactive)")
			}

			if err := checkEnvironment(cmd.OutOrStdout()); err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("🚀 Creating project '%s' using template '%s'\n", projectName, templateName)
			fmt.Println("───────────────────────────────────────────────────────")

			generator := template.NewGenerator(manager)
			if err := generator.Generate(projectName, templateName); err != nil {
				return err
			}

			showNextSteps(projectName)
			return nil
		},
	}

	templateName = "basic"

	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name for the generated application")
	cmd.Flags().StringVarP(&templateName, "template", "t", templateName, "Template to use when generating the project")
	cmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List available templates")
	cmd.Flags().StringVar(&installTarget, "install", "", "Install template from URL or local path")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")
	cmd.Flags().BoolVar(&versionFlag, "version", false, "Show the generator version")
	cmd.Flags().SortFlags = false

	return cmd
}

func listAvailableTemplates(manager *template.Manager) {
	templates := manager.ListTemplates()

	if len(templates) == 0 {
		fmt.Println("❌ No templates available.")
		return
	}

	printWelcomeBanner()
	fmt.Println("📋 Available templates:")
	fmt.Println("───────────────────────────────────────────────────────")

	for _, tmpl := range templates {
		fmt.Printf("📦 %s (%s)\n", tmpl.DisplayName, tmpl.Name)
		fmt.Printf("   %s\n", tmpl.Description)
		fmt.Printf("   Version: %s | Source: %s\n", tmpl.Version, tmpl.Source)
		if len(tmpl.Tags) > 0 {
			fmt.Printf("   🏷️  %s\n", strings.Join(tmpl.Tags, ", "))
		}
		fmt.Println()
	}
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Println()
}

func runInteractiveMode(manager *template.Manager) error {
	printWelcomeBanner()

	reader := bufio.NewReader(os.Stdin)

	templates := manager.ListTemplates()
	if len(templates) == 0 {
		return fmt.Errorf("no templates available")
	}

	fmt.Println("Available templates:")
	for i, tmpl := range templates {
		fmt.Printf("%d) %s - %s\n", i+1, tmpl.DisplayName, tmpl.Description)
	}

	var choice int
	for {
		fmt.Printf("\nSelect template (1-%d): ", len(templates))
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("Please enter a number.")
			continue
		}

		value, err := strconv.Atoi(input)
		if err != nil || value < 1 || value > len(templates) {
			fmt.Println("Invalid selection. Try again.")
			continue
		}

		choice = value
		break
	}

	selectedTemplate := templates[choice-1]

	var projectName string
	for {
		fmt.Print("Enter project name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read project name: %w", err)
		}

		projectName = strings.TrimSpace(input)
		if projectName == "" {
			fmt.Println("Project name cannot be empty. Please try again.")
			continue
		}

		break
	}

	if err := checkEnvironment(os.Stdout); err != nil {
		return err
	}

	generator := template.NewGenerator(manager)
	if err := generator.Generate(projectName, selectedTemplate.Name); err != nil {
		return err
	}

	showNextSteps(projectName)
	return nil
}

func showNextSteps(projectName string) {
	fmt.Println()
	fmt.Println("✨ Project created successfully!")
	fmt.Println()
	fmt.Println("📝 Next steps:")
	fmt.Printf("   cd %s\n", projectName)
	fmt.Println("   make install    # Install dependencies")
	fmt.Println("   make dev        # Start development servers")
	fmt.Println("   make build      # Build for production")
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Println("✨ Ready! Happy coding!")
	fmt.Println()
}

func checkEnvironment(out io.Writer) error {
	fmt.Fprintln(out, "🔄 Checking environment prerequisites...")
	if err := ensureTool("go", "Install Go from https://go.dev/dl/", out); err != nil {
		return err
	}
	if err := ensureTool("node", "Install Node.js from https://nodejs.org/", out); err != nil {
		return err
	}
	fmt.Fprintln(out, "✅ Environment ready")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	return nil
}

func ensureTool(name, hint string, out io.Writer) error {
	fmt.Fprintf(out, "   • %s: ", name)
	if _, err := exec.LookPath(name); err != nil {
		fmt.Fprintln(out, "❌ missing")
		return fmt.Errorf("%s executable not found in PATH. %s", name, hint)
	}
	fmt.Fprintln(out, "✅")
	return nil
}
