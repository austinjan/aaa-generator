package template

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	manager *Manager
}

func NewGenerator(manager *Manager) *Generator {
	return &Generator{manager: manager}
}

func (g *Generator) Generate(projectName, templateName string) error {
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectName)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check project directory: %w", err)
	}

	tmpl, err := g.manager.GetTemplate(templateName)
	if err != nil {
		return err
	}

	vars, err := g.collectVariables(tmpl.Config, projectName)
	if err != nil {
		return err
	}

	fmt.Println("🔄 Creating project directory...")
	if err := os.MkdirAll(projectName, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}
	fmt.Println("✅ Project directory created")

	fmt.Println("🔄 Generating project files...")
	if err := g.generateFiles(tmpl, projectName, vars); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	fmt.Println("✅ Project files generated")

	fmt.Println("🔄 Running post-generation commands...")
	if err := g.runPostCommands(tmpl.Config, projectName, vars); err != nil {
		return fmt.Errorf("failed to run post commands: %w", err)
	}
	fmt.Println("✅ Post-generation commands completed")

	return nil
}

func (g *Generator) collectVariables(config *TemplateConfig, projectName string) (map[string]interface{}, error) {
	vars := map[string]interface{}{
		"ProjectName": projectName,
		"ModuleName":  projectName,
	}

	if config == nil {
		return vars, nil
	}

	reader := bufio.NewReader(os.Stdin)

	for _, variable := range config.Variables {
		if _, exists := vars[variable.Name]; exists {
			continue
		}

		value := variable.Default

		if variable.Required && strings.TrimSpace(value) == "" {
			var err error
			value, err = g.promptForVariable(reader, variable)
			if err != nil {
				return nil, err
			}
		}

		if variable.Type == "select" && len(variable.Options) > 0 {
			if !contains(variable.Options, value) {
				return nil, fmt.Errorf("invalid value '%s' for variable '%s'. valid options: %v", value, variable.Name, variable.Options)
			}
		}

		vars[variable.Name] = value
	}

	return vars, nil
}

func (g *Generator) promptForVariable(reader *bufio.Reader, variable TemplateVar) (string, error) {
	for {
		fmt.Printf("Enter %s", variable.Name)
		if variable.Description != "" {
			fmt.Printf(" (%s)", variable.Description)
		}
		fmt.Print(": ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		value := strings.TrimSpace(input)
		if value == "" {
			fmt.Println("Value cannot be empty. Please try again.")
			continue
		}
		return value, nil
	}
}

func (g *Generator) generateFiles(tmpl *Template, projectName string, vars map[string]interface{}) error {
	useRules := tmpl.Config != nil && len(tmpl.Config.Files) > 0

	return fs.WalkDir(tmpl.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if path == "template.yaml" {
			return nil
		}

		relativePath := path
		matched := false
		if useRules {
			if mapped, ok := mapTargetPath(tmpl.Config.Files, path); ok {
				relativePath = mapped
				matched = true
			}
		}
		if useRules && !matched {
			return nil
		}

		relativePath = filepath.FromSlash(relativePath)
		targetPath := filepath.Join(projectName, relativePath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		content, readErr := fs.ReadFile(tmpl.Files, path)
		if readErr != nil {
			return readErr
		}

		if strings.HasSuffix(path, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
			return g.processTemplate(content, targetPath, vars)
		}

		return os.WriteFile(targetPath, content, 0o644)
	})
}

func (g *Generator) processTemplate(content []byte, targetPath string, vars map[string]interface{}) error {
	tmpl, err := template.New("template").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", targetPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, vars); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", targetPath, err)
	}

	return nil
}

func (g *Generator) runPostCommands(config *TemplateConfig, projectName string, vars map[string]interface{}) error {
	if config == nil || len(config.PostGenerate) == 0 {
		return nil
	}

	for _, command := range config.PostGenerate {
		cmdStr := g.processCommandTemplate(command.Command, vars)
		workDir := filepath.Join(projectName, command.WorkDir)
		if workDir == "" {
			workDir = projectName
		}

		fmt.Printf("   • Running: %s\n", cmdStr)

		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("   ⚠️  Warning: command failed: %s\n", cmdStr)
		}
	}

	return nil
}

func (g *Generator) processCommandTemplate(command string, vars map[string]interface{}) string {
	tmpl, err := template.New("command").Parse(command)
	if err != nil {
		return command
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, vars); err != nil {
		return command
	}

	return buf.String()
}

func mapTargetPath(rules []FileRule, path string) (string, bool) {
	normalized := strings.TrimPrefix(path, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "" {
		return "", true
	}

	for _, rule := range rules {
		src := strings.TrimSpace(rule.Source)
		if src == "" {
			continue
		}
		src = strings.TrimPrefix(src, "./")
		src = strings.TrimPrefix(src, "/")

		ruleType := strings.TrimSpace(rule.Type)
		if ruleType == "" {
			ruleType = "directory"
		}

		switch ruleType {
		case "directory":
			src = strings.TrimSuffix(src, "/")
			if src == "" {
				return joinRuleTarget(rule.Target, normalized), true
			}
			if normalized == src {
				return joinRuleTarget(rule.Target, ""), true
			}
			if strings.HasPrefix(normalized, src+"/") {
				rel := strings.TrimPrefix(normalized, src+"/")
				return joinRuleTarget(rule.Target, rel), true
			}
		case "file":
			src = strings.TrimSuffix(src, "/")
			if normalized == src {
				rel := strings.TrimSpace(rule.Target)
				if rel == "" {
					rel = filepath.Base(src)
				}
				rel = strings.TrimPrefix(rel, "./")
				return rel, true
			}
		}
	}

	return normalized, false
}

func joinRuleTarget(target, rel string) string {
	base := strings.TrimSpace(target)
	base = strings.Trim(base, "/")
	if base == "." {
		base = ""
	}

	if rel == "" {
		return base
	}

	if base == "" {
		return rel
	}

	return base + "/" + rel
}

func contains(options []string, value string) bool {
	for _, option := range options {
		if option == value {
			return true
		}
	}
	return false
}
