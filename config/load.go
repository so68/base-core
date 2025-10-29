package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/so68/utils/logger"
	"github.com/spf13/viper"
)

// LoadConfig 从指定路径加载配置文件
func LoadConfig(configPath string) (*AppConfig, error) {
	// 创建默认配置
	config := DefaultAppConfig()

	// 如果配置文件路径为空，尝试从环境变量或默认位置查找
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，返回默认配置
		config.SetDefaults()
		return config, nil
	}

	// 设置 viper 配置
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将配置绑定到结构体
	if err := viper.Unmarshal(config); err != nil {
		return config, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值（只设置未配置的字段）
	config.SetDefaults()

	return config, nil
}

// LoadConfigWithoutDefaults 从指定路径加载配置文件，不设置默认值
func LoadConfigWithoutDefaults(configPath string) (*AppConfig, error) {
	// 创建空配置，但初始化嵌套结构
	config := &AppConfig{
		Cors:      &CorsConfig{},
		JWT:       &JWTConfig{},
		RateLimit: &RateLimitConfig{},
		Logger:    logger.DefaultConfig(),
		Cache:     &CacheConfig{},
		Database:  &DatabaseConfig{},
	}

	// 如果配置文件路径为空，尝试从环境变量或默认位置查找
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，返回默认配置
		return DefaultAppConfig(), nil
	}

	// 创建新的 viper 实例
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return DefaultAppConfig(), fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将配置绑定到结构体
	if err := v.Unmarshal(config); err != nil {
		return DefaultAppConfig(), fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config, nil
}

// LoadConfigFromBytesWithoutDefaults 从字节数据加载配置，不设置默认值
func LoadConfigFromBytesWithoutDefaults(data []byte) (*AppConfig, error) {
	// 创建空配置，但初始化嵌套结构
	config := &AppConfig{
		Cors:      &CorsConfig{},
		JWT:       &JWTConfig{},
		RateLimit: &RateLimitConfig{},
		Logger:    logger.DefaultConfig(),
		Cache:     &CacheConfig{},
		Database:  &DatabaseConfig{},
	}

	// 创建新的 viper 实例
	v := viper.New()
	v.SetConfigType("yaml")

	// 从字节数据读取配置
	if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
		return DefaultAppConfig(), fmt.Errorf("读取配置数据失败: %w", err)
	}

	// 将配置绑定到结构体
	if err := v.Unmarshal(config); err != nil {
		return DefaultAppConfig(), fmt.Errorf("解析配置数据失败: %w", err)
	}

	return config, nil
}

// LoadConfigWithEnv 从环境变量和配置文件加载配置
func LoadConfigWithEnv(configPath string) (*AppConfig, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 从环境变量覆盖配置
	overrideFromEnv(config)

	return config, nil
}

// LoadConfigFromBytes 从字节数据加载配置
func LoadConfigFromBytes(data []byte) (*AppConfig, error) {
	// 创建默认配置
	config := DefaultAppConfig()

	// 设置 viper 配置
	viper.SetConfigType("yaml")

	// 从字节数据读取配置
	if err := viper.ReadConfig(strings.NewReader(string(data))); err != nil {
		return config, fmt.Errorf("读取配置数据失败: %w", err)
	}

	// 将配置绑定到结构体
	if err := viper.Unmarshal(config); err != nil {
		return config, fmt.Errorf("解析配置数据失败: %w", err)
	}

	// 设置默认值
	config.SetDefaults()

	return config, nil
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	// 按优先级查找配置文件
	possiblePaths := []string{
		"./config.yaml",
		"./config.yml",
		"./config/config.yaml",
		"./config/config.yml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 如果都没找到，返回默认路径
	return "./config.yaml"
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(config *AppConfig) {
	// 应用基础配置
	if val := os.Getenv("APP_NAME"); val != "" {
		config.Name = val
	}
	if val := os.Getenv("APP_HOST"); val != "" {
		config.Host = val
	}
	if val := os.Getenv("APP_PORT"); val != "" {
		if _, err := fmt.Sscanf(val, "%d", &config.Port); err == nil {
			// 成功解析端口
		}
	}
	if val := os.Getenv("APP_DEBUG"); val != "" {
		config.Debug = val == "true" || val == "1"
	}
	if val := os.Getenv("APP_READ_TIMEOUT"); val != "" {
		config.ReadTimeout = val
	}
	if val := os.Getenv("APP_WRITE_TIMEOUT"); val != "" {
		config.WriteTimeout = val
	}
	if val := os.Getenv("APP_IDLE_TIMEOUT"); val != "" {
		config.IdleTimeout = val
	}
	if val := os.Getenv("APP_MAX_HEADER"); val != "" {
		if _, err := fmt.Sscanf(val, "%d", &config.MaxHeader); err == nil {
			// 成功解析最大请求头
		}
	}

	// JWT 配置
	if config.JWT != nil {
		if val := os.Getenv("APP_JWT_SECRET_KEY"); val != "" {
			config.JWT.SecretKey = val
		}
		if val := os.Getenv("APP_JWT_EXPIRES_IN"); val != "" {
			if _, err := fmt.Sscanf(val, "%d", &config.JWT.ExpiresIn); err == nil {
				// 成功解析过期时间
			}
		}
		if val := os.Getenv("APP_JWT_HEADER_NAME"); val != "" {
			config.JWT.HeaderName = val
		}
		if val := os.Getenv("APP_JWT_SCHEME"); val != "" {
			config.JWT.Scheme = val
		}
	}

	// 数据库配置
	if config.Database != nil {
		if val := os.Getenv("APP_DATABASE_HOST"); val != "" {
			config.Database.Host = val
		}
		if val := os.Getenv("APP_DATABASE_PORT"); val != "" {
			if _, err := fmt.Sscanf(val, "%d", &config.Database.Port); err == nil {
				// 成功解析端口
			}
		}
		if val := os.Getenv("APP_DATABASE_USERNAME"); val != "" {
			config.Database.Username = val
		}
		if val := os.Getenv("APP_DATABASE_PASSWORD"); val != "" {
			config.Database.Password = val
		}
		if val := os.Getenv("APP_DATABASE_DATABASE"); val != "" {
			config.Database.Database = val
		}
	}

	// 缓存配置
	if config.Cache != nil {
		if val := os.Getenv("APP_CACHE_HOST"); val != "" {
			config.Cache.Host = val
		}
		if val := os.Getenv("APP_CACHE_PORT"); val != "" {
			if _, err := fmt.Sscanf(val, "%d", &config.Cache.Port); err == nil {
				// 成功解析端口
			}
		}
		if val := os.Getenv("APP_CACHE_PASSWORD"); val != "" {
			config.Cache.Password = val
		}
	}
}

// ValidateConfig 验证配置的有效性
func ValidateConfig(config *AppConfig) error {
	if config == nil {
		return fmt.Errorf("配置对象不能为空")
	}

	// 验证端口范围
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d", config.Port)
	}

	// 验证 JWT 配置
	if config.JWT != nil {
		if config.JWT.SecretKey == "" {
			return fmt.Errorf("JWT 密钥不能为空")
		}
		if config.JWT.ExpiresIn <= 0 {
			return fmt.Errorf("JWT 过期时间必须大于 0")
		}
	}

	// 验证数据库配置
	if config.Database != nil {
		if config.Database.Host == "" {
			return fmt.Errorf("数据库主机不能为空")
		}
		if config.Database.Port <= 0 || config.Database.Port > 65535 {
			return fmt.Errorf("无效的数据库端口号: %d", config.Database.Port)
		}
		if config.Database.Username == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
		if config.Database.Database == "" {
			return fmt.Errorf("数据库名称不能为空")
		}
	}

	// 验证缓存配置
	if config.Cache != nil {
		if config.Cache.Host == "" {
			return fmt.Errorf("缓存主机不能为空")
		}
		if config.Cache.Port <= 0 || config.Cache.Port > 65535 {
			return fmt.Errorf("无效的缓存端口号: %d", config.Cache.Port)
		}
	}

	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *AppConfig, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建新的 viper 实例
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	// 直接设置配置值
	v.Set("name", config.Name)
	v.Set("host", config.Host)
	v.Set("port", config.Port)
	v.Set("debug", config.Debug)
	v.Set("read_timeout", config.ReadTimeout)
	v.Set("write_timeout", config.WriteTimeout)
	v.Set("idle_timeout", config.IdleTimeout)
	v.Set("max_header", config.MaxHeader)

	// 设置子配置
	if config.Cors != nil {
		v.Set("cors", config.Cors)
	}
	if config.JWT != nil {
		v.Set("jwt", config.JWT)
	}
	if config.RateLimit != nil {
		v.Set("rate_limit", config.RateLimit)
	}
	if config.Logger != nil {
		v.Set("logger", config.Logger)
	}
	if config.Cache != nil {
		v.Set("cache", config.Cache)
	}
	if config.Database != nil {
		v.Set("database", config.Database)
	}

	// 写入文件
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
