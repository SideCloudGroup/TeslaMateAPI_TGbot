package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config 全局配置结构
type Config struct {
	Telegram  TelegramConfig  `toml:"telegram"`
	TeslaMate TeslaMateConfig `toml:"teslamate"`
}

// TelegramConfig Telegram Bot配置
type TelegramConfig struct {
	BotToken         string  `toml:"bot_token"`
	WhitelistChatIDs []int64 `toml:"whitelist_chat_ids"`
	APIEndpoint      string  `toml:"api_endpoint"` // 自定义API端点（可选）
}

// TeslaMateConfig TeslaMate API配置
type TeslaMateConfig struct {
	APIURL  string            `toml:"api_url"`
	APIKey  string            `toml:"api_key"`
	CarID   int               `toml:"car_id"`
	Timeout int               `toml:"timeout"`
	Headers map[string]string `toml:"headers"` // 自定义请求头（可选）
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	var config Config

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析TOML
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证必填项
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate 验证配置项
func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram.bot_token 不能为空")
	}
	if len(c.Telegram.WhitelistChatIDs) == 0 {
		return fmt.Errorf("telegram.whitelist_chat_ids 不能为空")
	}
	// api_endpoint为可选项，如果为空则使用默认Telegram API
	if c.TeslaMate.APIURL == "" {
		return fmt.Errorf("teslamate.api_url 不能为空")
	}
	// api_key 可选
	if c.TeslaMate.CarID <= 0 {
		return fmt.Errorf("teslamate.car_id 必须大于0")
	}
	if c.TeslaMate.Timeout <= 0 {
		c.TeslaMate.Timeout = 30 // 默认30秒
	}
	return nil
}
