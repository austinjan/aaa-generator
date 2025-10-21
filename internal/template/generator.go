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
	// 檢查項目目錄是否已存在
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	// 獲取模板
	tmpl, err := g.manager.GetTemplate(templateName)
	if err != nil {
		return err
	}

	fmt.Printf("🚀 Creating project '%s' using template '%s'...\n", projectName, tmpl.Config.DisplayName)

	// 收集變數
	vars, err := g.collectVariables(tmpl.Config, projectName)
	if err != nil {
		return err
	}

	// 創建項目目錄
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// 生成文件
	if err := g.generateFiles(tmpl, projectName, vars); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// 執行後處理指令
	if err := g.runPostCommands(tmpl.Config, projectName, vars); err != nil {
		return fmt.Errorf("failed to run post commands: %w", err)
	}

	return nil
}

func (g *Generator) collectVariables(config *TemplateConfig, projectName string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})

	// 設置基本變數
	vars["ProjectName"] = projectName
	vars["ModuleName"] = fmt.Sprintf("github.com/yourname/%s", projectName)

	// 處理模板定義的變數
	for _, variable := range config.Variables {
		// 跳過已設置的變數
		if _, exists := vars[variable.Name]; exists {
			continue
		}

		var value string
		if variable.Default != "" {
			value = variable.Default
		}

		// 如果變數是必需的且沒有默認值，提示用戶輸入
		if variable.Required && value == "" {
			fmt.Printf("Enter %s", variable.Name)
			if variable.Description != "" {
				fmt.Printf(" (%s)", variable.Description)
			}
			fmt.Print(": ")

			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			value = strings.TrimSpace(input)

			if value == "" {
				return nil, fmt.Errorf("required variable '%s' cannot be empty", variable.Name)
			}
		}

		// 對於選擇類型的變數，驗證值
		if variable.Type == "select" && len(variable.Options) > 0 {
			validOption := false
			for _, option := range variable.Options {
				if value == option {
					validOption = true
					break
				}
			}
			if !validOption {
				return nil, fmt.Errorf("invalid value '%s' for variable '%s'. Valid options: %v", value, variable.Name, variable.Options)
			}
		}

		vars[variable.Name] = value
	}

	return vars, nil
}

func (g *Generator) generateFiles(tmpl *Template, projectName string, vars map[string]interface{}) error {
	return fs.WalkDir(tmpl.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳過模板配置文件
		if path == "template.yaml" {
			return nil
		}

		targetPath := filepath.Join(projectName, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// 讀取文件內容
		content, err := fs.ReadFile(tmpl.Files, path)
		if err != nil {
			return err
		}

		// 如果是模板文件，處理模板
		if strings.HasSuffix(path, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
			return g.processTemplate(content, targetPath, vars)
		}

		// 直接複製非模板文件
		return os.WriteFile(targetPath, content, 0644)
	})
}

func (g *Generator) processTemplate(content []byte, targetPath string, vars map[string]interface{}) error {
	tmpl, err := template.New("template").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", targetPath, err)
	}

	// 確保目標目錄存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
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
	if len(config.PostGenerate) == 0 {
		return nil
	}

	fmt.Println("📦 Running post-generation commands...")

	for _, command := range config.PostGenerate {
		// 處理命令中的模板變數
		cmdStr := g.processCommandTemplate(command.Command, vars)
		workDir := filepath.Join(projectName, command.WorkDir)

		fmt.Printf("  Running: %s\n", cmdStr)

		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: Command failed: %s\n", cmdStr)
			// 不返回錯誤，繼續執行其他命令
		}
	}

	return nil
}

func (g *Generator) processCommandTemplate(command string, vars map[string]interface{}) string {
	tmpl, err := template.New("command").Parse(command)
	if err != nil {
		return command // 返回原始命令
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, vars); err != nil {
		return command // 返回原始命令
	}

	return buf.String()
}
