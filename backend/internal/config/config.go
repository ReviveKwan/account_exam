package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	CloudBase CloudBaseConfig
	JWT       JWTConfig
	WeChat    WeChatConfig
	Server    ServerConfig
}

// CloudBaseConfig 云开发自带文档库（HTTP API）
type CloudBaseConfig struct {
	EnvID    string
	APIKey   string
	Instance string
	Database string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type WeChatConfig struct {
	AppID  string
	Secret string
}

type ServerConfig struct {
	Port int
	Mode string
}

func Load() *Config {
	// 先加载文件中的环境变量（不覆盖已存在的系统环境变量）
	loadEnvFile(".env")
	loadEnvFile("config/.env")
	loadEnvFile("config/apikey") // 单行纯 Key 文件
	loadAPIKeyFile("config/apikey")

	return &Config{
		CloudBase: CloudBaseConfig{
			EnvID: getEnv("CLOUDBASE_ENV_ID", getEnv("TCB_ENV", "wechat-app-d6gjceiz966031f7f")),
			APIKey: stringsTrim(firstNonEmpty(
				os.Getenv("CLOUDBASE_APIKEY"),
				os.Getenv("CLOUDBASE_API_KEY"),
				os.Getenv("APIKEY"),
			)),
			Instance: getEnv("CLOUDBASE_INSTANCE", "(default)"),
			Database: getEnv("CLOUDBASE_DATABASE", "(default)"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "mobi-tou-exam-2025-secret-key"),
			ExpireHours: 168,
		},
		WeChat: WeChatConfig{
			AppID:  getEnv("WX_APPID", ""),
			Secret: getEnv("WX_SECRET", ""),
		},
		Server: ServerConfig{
			Port: resolvePort(),
			Mode: getEnv("GIN_MODE", "debug"),
		},
	}
}

func resolvePort() int {
	if p := getEnvInt("PORT", 0); p > 0 {
		return p
	}
	if p := getEnvInt("SERVER_PORT", 0); p > 0 {
		return p
	}
	return 80
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func stringsTrim(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, "\"'")
	return v
}

// loadEnvFile 解析 KEY=VALUE；已存在的环境变量不覆盖
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := stringsTrim(line[idx+1:])
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

// loadAPIKeyFile 若文件内容是纯 Key（不含 =），写入 CLOUDBASE_APIKEY
func loadAPIKeyFile(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := stringsTrim(string(b))
	if content == "" || strings.Contains(content, "=") {
		return
	}
	if os.Getenv("CLOUDBASE_APIKEY") == "" {
		_ = os.Setenv("CLOUDBASE_APIKEY", content)
	}
}
