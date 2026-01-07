package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"go_shell/shell"

	"github.com/chzyer/readline"
)

func main() {
	trie := shell.InitTrie()
	// 启动时只从 HISTFILE 加载一次历史记录
	shell.InitHistoryFile()

	// github.com/chzyer/readline 是一个 Go 语言的 readline 库
	// 它提供了类似 bash 的交互式命令行输入功能，包括：
	// - 行编辑（左右箭头键移动光标）
	// - 历史记录（上下箭头键浏览历史）
	// - 自动补全（TAB 键触发）
	// - 快捷键支持（Ctrl+C, Ctrl+D 等）
	//
	// NewEx 创建一个可配置的 readline 实例
	// Config 结构体包含各种配置选项
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",                        // 提示符，显示在每行输入前
		AutoComplete:    shell.CreateCompleter(trie), // 自动补全器，当用户按下 TAB 键时调用
		InterruptPrompt: "^C",                        // 当用户按下 Ctrl+C 时显示的提示
		EOFPrompt:       "exit",                      // 当用户按下 Ctrl+D (EOF) 时显示的提示
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close() // 关闭 readline 实例，释放资源

mainLoop:
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			// 这里就是“监听 EOF / Ctrl+D”的地方
			if err == io.EOF {
				shell.SaveCmdHistoryToEnvFile()
			}
			break
		}
		shell.HistoryCmdSlice = append(shell.HistoryCmdSlice, line)
		realCommand := strings.TrimLeft(line, " ")
		realCommand = strings.TrimRight(realCommand, " ")

		// 检查是否有管道操作符
		pipeParts := shell.ParsePipe(realCommand)
		if len(pipeParts) > 1 {
			// 有管道，执行管道逻辑
			shell.HandlePipeline(pipeParts)
			continue
		}

		cmdSlice := shell.ParseCommand(realCommand)
		if len(cmdSlice) == 0 {
			continue
		}

		// 解析重定向操作符
		actualCmdSlice, stdoutFile, stderrFile, stdappendFile, stderrappendFile := shell.ParseRedirect(cmdSlice)
		if len(actualCmdSlice) == 0 {
			continue
		}

		commandName := actualCmdSlice[0]

		// 如果有标准输出重定向，打开文件用于写入
		var outputFile *os.File
		if stdoutFile != "" {
			outputFile, err = os.Create(stdoutFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file %s: %v\n", stdoutFile, err)
				continue
			}
			defer outputFile.Close()
		}

		// 如果有标准错误重定向，打开文件用于写入错误
		var errorFile *os.File
		if stderrFile != "" {
			errorFile, err = os.Create(stderrFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file %s: %v\n", stderrFile, err)
				continue
			}
			defer errorFile.Close()
		}
		// 如果有标准追加重定向，打开文件用于追加写入
		var appendFile *os.File
		if stdappendFile != "" {
			appendFile, err = os.OpenFile(stdappendFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", stdappendFile, err)
				continue
			}
			defer appendFile.Close()
		}
		// 如果有标准错误追加重定向，打开文件用于追加写入
		var errAppendFile *os.File
		if stderrappendFile != "" {
			errAppendFile, err = os.OpenFile(stderrappendFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", stderrappendFile, err)
				continue
			}
			defer errAppendFile.Close()
		}

		// 对于 echo 命令，需要从原始命令中移除重定向部分
		if commandName == "echo" {
			// 从 realCommand 中移除重定向部分
			echoCmd := realCommand
			// 移除标准输出重定向
			if stdoutFile != "" {
				redirectPatterns := []string{" 1> " + stdoutFile, " > " + stdoutFile}
				for _, pattern := range redirectPatterns {
					if idx := strings.LastIndex(realCommand, pattern); idx != -1 {
						echoCmd = strings.TrimRight(realCommand[:idx], " ")
						break
					}
				}
			}
			// 移除标准错误重定向
			if stderrFile != "" {
				pattern := " 2> " + stderrFile
				if idx := strings.LastIndex(echoCmd, pattern); idx != -1 {
					echoCmd = strings.TrimRight(echoCmd[:idx], " ")
				}
			}
			// 移除标准追加重定向
			if stdappendFile != "" {
				redirectPatterns := []string{" 1>> " + stdappendFile, " >> " + stdappendFile}
				for _, pattern := range redirectPatterns {
					if idx := strings.LastIndex(echoCmd, pattern); idx != -1 {
						echoCmd = strings.TrimRight(echoCmd[:idx], " ")
						break
					}
				}
			}
			// 移除标准错误追加重定向
			if stderrappendFile != "" {
				pattern := " 2>> " + stderrappendFile
				if idx := strings.LastIndex(echoCmd, pattern); idx != -1 {
					echoCmd = strings.TrimRight(echoCmd[:idx], " ")
				}
			}
			if shell.HandleEcho(echoCmd, outputFile, appendFile) {
				continue
			}
		}
		switch commandName {
		case "exit":
			if shell.HandleExit() {
				break mainLoop
			}
		case "type":
			if shell.HandleType(actualCmdSlice, outputFile) {
				continue
			}
		case "pwd":
			if shell.HandlePwd(outputFile, errorFile, appendFile, errAppendFile) {
				continue
			}
		case "cd":
			if shell.HandleCD(actualCmdSlice, errorFile, errAppendFile) {
				continue
			}
		case "history":
			if shell.HandleHistory(actualCmdSlice, outputFile, appendFile) {
				continue
			}
		default:
			// 如果不是内置命令，尝试作为外部程序执行
			shell.HandleExternalCommand(commandName, actualCmdSlice, outputFile, errorFile, appendFile, errAppendFile)
		}
	}
}
