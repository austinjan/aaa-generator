package template

import "fmt"

type TemplateConfig struct {
	Name         string        `yaml:"name"`
	DisplayName  string        `yaml:"displayName"`
	Description  string        `yaml:"description"`
	Version      string        `yaml:"version"`
	Author       string        `yaml:"author"`
	Tags         []string      `yaml:"tags"`
	Variables    []TemplateVar `yaml:"variables"`
	Files        []FileRule    `yaml:"files"`
	PostGenerate []PostCommand `yaml:"postGenerate"`
}

type TemplateVar struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"` // string, int, bool, select
	Required    bool     `yaml:"required"`
	Default     string   `yaml:"default"`
	Options     []string `yaml:"options"`
	Description string   `yaml:"description"`
}

type FileRule struct {
	Source    string `yaml:"source"`
	Target    string `yaml:"target"`
	Type      string `yaml:"type"` // file, directory
	Condition string `yaml:"condition"`
}

type PostCommand struct {
	Command string `yaml:"command"`
	WorkDir string `yaml:"workDir"`
}

// 項目變數結構
type ProjectVars struct {
	ProjectName  string
	ModuleName   string
	Port         string
	DatabaseType string
	AuthProvider string
	// 可根據需要添加更多變數
}

// 獲取默認變數
func GetDefaultVars(projectName string) ProjectVars {
	return ProjectVars{
		ProjectName:  projectName,
		ModuleName:   fmt.Sprintf("github.com/yourname/%s", projectName),
		Port:         "8080",
		DatabaseType: "none",
		AuthProvider: "none",
	}
}
