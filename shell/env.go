package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go_shell/utils"
)

var ShellSlice = []string{"echo", "type", "exit", "pwd", "cd", "history"}
var HistoryCmdSlice = []string{}
var lastHistoryWrittenIndex = 0

func InitHistoryFile() {
	// 从环境变量HISTFILE中加载历史记录
	historyFilePath := os.Getenv("HISTFILE")
	if historyFilePath != "" {
		fileBytes, err := os.ReadFile(historyFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error read file %s\n", err)
		}
		fileString := string(fileBytes)
		spFileString := strings.Split(fileString, "\n")
		for i := 0; i < len(spFileString)-1; i++ { // 跳过最后一个空元素
			HistoryCmdSlice = append(HistoryCmdSlice, spFileString[i])
		}
		// 已有的历史记录来自文件，不应该在退出时再次写入
		lastHistoryWrittenIndex = len(HistoryCmdSlice)
	}
}

func SaveCmdHistoryToEnvFile() {
	historyFilePath := os.Getenv("HISTFILE")
	if historyFilePath != "" {
		if lastHistoryWrittenIndex < 0 || lastHistoryWrittenIndex > len(HistoryCmdSlice) {
			lastHistoryWrittenIndex = 0
		}
		appendFile, err := os.OpenFile(historyFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error open file %s\n", err)
			return
		}
		defer appendFile.Close()
		newEntries := HistoryCmdSlice[lastHistoryWrittenIndex:]
		if len(newEntries) > 0 {
			contents := strings.Join(newEntries, "\n") + "\n" // 加上尾部换行符
			_, err = appendFile.Write([]byte(contents))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error write file %s\n", err)
				return
			}
			lastHistoryWrittenIndex = len(HistoryCmdSlice)
		}
	}
}

func InitTrie() *utils.Trie {
	trie := utils.Constructor()
	// 插入内置命令
	trie.Insert("echo")
	trie.Insert("exit")
	trie.Insert("type")
	trie.Insert("pwd")
	trie.Insert("cd")

	// 扫描 PATH 环境变量中的可执行文件并插入到 Trie
	loadExecutablesFromPath(trie)

	return trie
}

// loadExecutablesFromPath 扫描 PATH 环境变量中的所有目录，找到可执行文件并插入到 Trie
func loadExecutablesFromPath(trie *utils.Trie) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return
	}

	// 分割 PATH 环境变量（在 Unix 系统中使用 :，在 Windows 中使用 ;）
	separator := string(os.PathListSeparator)

	dirs := strings.Split(pathEnv, separator)
	seen := make(map[string]bool) // 用于去重，避免重复插入相同的命令名

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// 读取目录内容
		entries, err := os.ReadDir(dir)
		if err != nil {
			// 如果目录不存在或无法读取，跳过
			continue
		}

		// 遍历目录中的文件
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()

			// 检查文件是否可执行
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// 检查文件是否可执行
			isExecutable := false
			if info.Mode().IsRegular() {
				// 在 Unix 系统上，检查执行权限位
				if info.Mode()&0111 != 0 {
					isExecutable = true
				}
				// 在 Windows 系统上，检查文件扩展名
				if os.PathSeparator == '\\' {
					ext := strings.ToLower(filepath.Ext(name))
					if ext == ".exe" || ext == ".bat" || ext == ".cmd" || ext == ".com" {
						isExecutable = true
					}
				}
			}

			if isExecutable {
				// 转换为小写以保持一致性（命令名通常不区分大小写 Linux区别大小写 Windows不区别大小写）
				// 在 Windows 上，移除扩展名（用户输入命令时通常不包含扩展名）
				lowerName := strings.ToLower(name)
				if os.PathSeparator == '\\' {
					lowerName = strings.TrimSuffix(lowerName, filepath.Ext(lowerName))
				}

				// 去重：如果已经插入过，跳过
				// Trie 现在支持所有字符（包括数字、下划线等），所以不需要字符验证
				if !seen[lowerName] && lowerName != "" {
					trie.Insert(lowerName)
					seen[lowerName] = true
				}
			}
		}
	}
}
