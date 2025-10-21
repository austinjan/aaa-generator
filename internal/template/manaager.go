package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed all:templates
var embeddedTemplates embed.FS // `all:` prefix ensures directories are included.

type Manager struct {
	localTemplates map[string]*Template
	userTemplates  map[string]*Template
}

type Template struct {
	Config    *TemplateConfig
	Files     fs.FS
	LocalPath string
}

type TemplateInfo struct {
	Name        string
	DisplayName string
	Description string
	Version     string
	Source      string
	Tags        []string
	URL         string
}

func NewManager() (*Manager, error) {
	manager := &Manager{
		localTemplates: make(map[string]*Template),
		userTemplates:  make(map[string]*Template),
	}

	// 載入內嵌模板
	if err := manager.loadEmbeddedTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load embedded templates: %w", err)
	}

	// 載入用戶自定義模板
	if err := manager.loadUserTemplates(); err != nil {
		// 非致命錯誤，僅記錄日誌
		fmt.Printf("Warning: Failed to load user templates: %v\n", err)
	}

	return manager, nil
}

func (m *Manager) loadEmbeddedTemplates() error {
	entries, err := embeddedTemplates.ReadDir("templates")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		templateName := entry.Name()
		configPath := fmt.Sprintf("templates/%s/template.yaml", templateName)

		configData, err := embeddedTemplates.ReadFile(configPath)
		if err != nil {
			fmt.Printf("Warning: Failed to read config for template %s: %v\n", templateName, err)
			continue
		}

		var config TemplateConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			fmt.Printf("Warning: Failed to parse config for template %s: %v\n", templateName, err)
			continue
		}

		templateFS, err := fs.Sub(embeddedTemplates, fmt.Sprintf("templates/%s", templateName))
		if err != nil {
			fmt.Printf("Warning: Failed to create sub-filesystem for template %s: %v\n", templateName, err)
			continue
		}

		m.localTemplates[templateName] = &Template{
			Config: &config,
			Files:  templateFS,
		}
	}

	return nil
}

func (m *Manager) loadUserTemplates() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	templatesDir := filepath.Join(homeDir, ".go-react-generator", "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// 目錄不存在，創建它
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			return err
		}
		return nil
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		templateName := entry.Name()
		templatePath := filepath.Join(templatesDir, templateName)
		configPath := filepath.Join(templatePath, "template.yaml")

		configData, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("Warning: Failed to read config for user template %s: %v\n", templateName, err)
			continue
		}

		var config TemplateConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			fmt.Printf("Warning: Failed to parse config for user template %s: %v\n", templateName, err)
			continue
		}

		m.userTemplates[templateName] = &Template{
			Config:    &config,
			Files:     os.DirFS(templatePath),
			LocalPath: templatePath,
		}
	}

	return nil
}

func (m *Manager) ListTemplates() []TemplateInfo {
	var templates []TemplateInfo

	// 本地模板
	for _, tmpl := range m.localTemplates {
		templates = append(templates, TemplateInfo{
			Name:        tmpl.Config.Name,
			DisplayName: tmpl.Config.DisplayName,
			Description: tmpl.Config.Description,
			Version:     tmpl.Config.Version,
			Source:      "built-in",
			Tags:        tmpl.Config.Tags,
		})
	}

	// 用戶模板
	for _, tmpl := range m.userTemplates {
		templates = append(templates, TemplateInfo{
			Name:        tmpl.Config.Name,
			DisplayName: tmpl.Config.DisplayName,
			Description: tmpl.Config.Description,
			Version:     tmpl.Config.Version,
			Source:      "user",
			Tags:        tmpl.Config.Tags,
		})
	}

	return templates
}

func (m *Manager) GetTemplate(name string) (*Template, error) {
	// 優先級: 用戶模板 > 本地模板
	if tmpl, exists := m.userTemplates[name]; exists {
		return tmpl, nil
	}

	if tmpl, exists := m.localTemplates[name]; exists {
		return tmpl, nil
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

func (m *Manager) InstallTemplate(source string) error {
	if strings.HasPrefix(source, "http") {
		return m.installRemoteTemplate(source)
	}
	return m.installLocalTemplate(source)
}

func (m *Manager) installLocalTemplate(sourcePath string) error {
	// 檢查源路徑是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// 讀取模板配置
	configPath := filepath.Join(sourcePath, "template.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read template config: %w", err)
	}

	var config TemplateConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse template config: %w", err)
	}

	// 創建用戶模板目錄
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	userTemplatesDir := filepath.Join(homeDir, ".go-react-generator", "templates")
	targetPath := filepath.Join(userTemplatesDir, config.Name)

	// 複製模板文件
	if err := copyDir(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	fmt.Printf("✅ Template '%s' installed successfully!\n", config.Name)
	return nil
}

func (m *Manager) installRemoteTemplate(url string) error {
	// TODO: 實現從 Git URL 下載模板
	return fmt.Errorf("remote template installation not implemented yet")
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 複製文件
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = srcFile.WriteTo(dstFile)
		return err
	})
}
