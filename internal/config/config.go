// Package config 负责冰点咖啡的启动参数加载。
// 支持 config.yaml / 环境变量 / 命令行参数三层覆盖。
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Config 冰点咖啡的完整配置。
type Config struct {
	// AI 配置
	AI AIConfig `mapstructure:"ai"`

	// 服务器配置
	Server ServerConfig `mapstructure:"server"`

	// 插件配置
	Plugin PluginConfig `mapstructure:"plugin"`

	// 数据库配置
	DB DBConfig `mapstructure:"db"`

	// 日志配置
	Log LogConfig `mapstructure:"log"`
}

// AIConfig AI 模型配置。
type AIConfig struct {
	// API 地址（OpenAI 兼容）
	// 例如: https://api.openai.com/v1
	//       https://api.deepseek.com/v1
	//       https://dashscope.aliyuncs.com/compatible-mode/v1
	BaseURL string `mapstructure:"base_url"`

	// API Key
	APIKey string `mapstructure:"api_key"`

	// 模型名称
	Model string `mapstructure:"model"`

	// 温度参数（0-2）
	Temperature float64 `mapstructure:"temperature"`

	// 最大 token 数
	MaxTokens int `mapstructure:"max_tokens"`
}

// ServerConfig 租赁服连接配置。
type ServerConfig struct {
	// 服务器地址
	Address string `mapstructure:"address"`

	// 玩家名称
	PlayerName string `mapstructure:"player_name"`

	// FB Master Token
	FBToken string `mapstructure:"fb_token"`

	// 连接超时（秒）
	Timeout int `mapstructure:"timeout"`
}

// PluginConfig 插件配置。
type PluginConfig struct {
	// 插件目录
	Dir string `mapstructure:"dir"`

	// HTTP RPC 端口（0=禁用）
	HTTPPort int `mapstructure:"http_port"`
}

// DBConfig 数据库配置。
type DBConfig struct {
	// 数据库路径（支持 ~ 展开）
	Path string `mapstructure:"path"`
}

// LogConfig 日志配置。
type LogConfig struct {
	// 日志文件路径（空=stdout）
	File string `mapstructure:"file"`

	// 日志级别：debug / info / warn / error
	Level string `mapstructure:"level"`
}

// Load 加载配置。
// 优先级：命令行 flags > 环境变量 > config.yaml。
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 环境变量前缀
	v.SetEnvPrefix("IC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 如果指定了配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认在 ~/.icepoint/config.yaml
		home, err := homedir.Expand("~/.icepoint")
		if err == nil {
			os.MkdirAll(home, 0755)
			v.AddConfigPath(home)
			v.SetConfigName("config")
			v.SetConfigType("yaml")
		}
	}

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 文件不存在，用默认值+环境变量
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 环境变量覆盖
	applyEnvOverrides(&cfg)

	// 校验
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// AI 默认值
	v.SetDefault("ai.base_url", "https://api.openai.com/v1")
	v.SetDefault("ai.model", "gpt-4o-mini")
	v.SetDefault("ai.temperature", 0.7)
	v.SetDefault("ai.max_tokens", 4096)

	// 服务器默认值
	v.SetDefault("server.timeout", 30)

	// 插件默认值
	v.SetDefault("plugin.dir", "./plugins")
	v.SetDefault("plugin.http_port", 0)

	// 数据库默认值
	v.SetDefault("db.path", "~/.icepoint/data.db")

	// 日志默认值
	v.SetDefault("log.level", "info")
}

func applyEnvOverrides(cfg *Config) {
	// AI
	if v := os.Getenv("IC_AI_BASE_URL"); v != "" {
		cfg.AI.BaseURL = v
	}
	if v := os.Getenv("IC_AI_API_KEY"); v != "" {
		cfg.AI.APIKey = v
	}
	if v := os.Getenv("IC_AI_MODEL"); v != "" {
		cfg.AI.Model = v
	}

	// 服务器
	if v := os.Getenv("IC_SERVER_ADDRESS"); v != "" {
		cfg.Server.Address = v
	}
	if v := os.Getenv("IC_SERVER_FB_TOKEN"); v != "" {
		cfg.Server.FBToken = v
	}
	if v := os.Getenv("IC_SERVER_PLAYER_NAME"); v != "" {
		cfg.Server.PlayerName = v
	}
}

// Validate 校验配置是否合法。
func (c *Config) Validate() error {
	if c.AI.BaseURL == "" {
		return fmt.Errorf("ai.base_url 不能为空")
	}
	if c.AI.APIKey == "" {
		return fmt.Errorf("ai.api_key 不能为空")
	}
	if c.Server.Address == "" {
		return fmt.Errorf("server.address 不能为空")
	}
	if c.Server.FBToken == "" {
		return fmt.Errorf("server.fb_token 不能为空")
	}
	return nil
}
