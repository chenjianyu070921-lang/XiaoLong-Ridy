package svc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadDotEnvFile 读取本地 .env 文件，并将尚未存在的配置写入当前进程环境。
// 已由操作系统或部署平台注入的同名变量优先级更高，避免本地文件覆盖线上配置。
func loadDotEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("打开本地环境配置失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		if !found || key == "" {
			return fmt.Errorf("本地环境配置第 %d 行格式错误", lineNumber)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("设置本地环境配置 %s 失败: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取本地环境配置失败: %w", err)
	}
	return nil
}
