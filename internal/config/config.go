package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	API     APIConfig     `mapstructure:"api"`
	Tools   ToolsConfig   `mapstructure:"tools"`
	DAG     DAGConfig     `mapstructure:"dag"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// APIConfig API配置
type APIConfig struct {
	OpenAIKey string `mapstructure:"openai_key"`
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	Timeout   int    `mapstructure:"timeout"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	Enabled        []string              `mapstructure:"enabled"`
	WriteCode      WriteCodeConfig       `mapstructure:"write_code"`
	ReadFile       ReadFileConfig        `mapstructure:"read_file"`
	RecognizeImage RecognizeImageConfig  `mapstructure:"recognize_image"`
}

// WriteCodeConfig 代码写入工具配置
type WriteCodeConfig struct {
	MaxLines           int      `mapstructure:"max_lines"`
	SupportedLanguages []string `mapstructure:"supported_languages"`
}

// ReadFileConfig 文件读取工具配置
type ReadFileConfig struct {
	MaxSizeMB         int      `mapstructure:"max_size_mb"`
	AllowedExtensions []string `mapstructure:"allowed_extensions"`
}

// RecognizeImageConfig 图片识别工具配置
type RecognizeImageConfig struct {
	MaxSizeMB        int      `mapstructure:"max_size_mb"`
	SupportedFormats []string `mapstructure:"supported_formats"`
}

// DAGConfig DAG思考引擎配置
type DAGConfig struct {
	MaxDepth      int  `mapstructure:"max_depth"`
	ParallelNodes int  `mapstructure:"parallel_nodes"`
	Timeout       int  `mapstructure:"timeout"`
	Verbose       bool `mapstructure:"verbose"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
	Format string `mapstructure:"format"`
}

var globalConfig *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认配置文件路径
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")

		// 获取可执行文件目录
		if ex, err := os.Executable(); err == nil {
			v.AddConfigPath(filepath.Dir(ex))
		}
	}

	// 环境变量支持
	v.SetEnvPrefix("AGENT")
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必要配置
	if cfg.API.OpenAIKey == "" {
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			cfg.API.OpenAIKey = key
		} else {
			return nil, fmt.Errorf("未配置API Key，请在配置文件中设置api.openai_key或设置环境变量OPENAI_API_KEY")
		}
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}
