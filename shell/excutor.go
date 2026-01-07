package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"go_shell/utils"
)

// 处理CD命令  cmdSlice 0:cd 1:dir
func HandleCD(cmdSlice []string, errorFile *os.File, errorAppendFile *os.File) bool {
	errorWriter := os.Stderr
	if errorFile != nil {
		errorWriter = errorFile
	} else if errorAppendFile != nil {
		errorWriter = errorAppendFile
	}

	if len(cmdSlice) < 2 {
		return false
	}

	targetPath := cmdSlice[1]
	if targetPath == "~" {
		homePath := os.Getenv("HOME")
		if homePath == "" {
			fmt.Fprintf(errorWriter, "cd: HOME not set\n")
			return false
		}
		err := os.Chdir(homePath)
		if err != nil {
			fmt.Fprintf(errorWriter, "cd: %s: No such file or directory\n", homePath)
			return false
		}
		return true
	}

	// 判断是绝对路径还是相对路径
	var fullPath string
	if filepath.IsAbs(targetPath) {
		// 绝对路径：直接使用
		fullPath = targetPath
	} else {
		// 相对路径：需要与当前工作目录组合
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(errorWriter, "Error getting current directory: %v\n", err)
			return false
		}
		fullPath = filepath.Join(wd, targetPath)
	}

	// 检查目录是否存在
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		// 目录不存在，打印错误信息
		fmt.Fprintf(errorWriter, "cd: %s: No such file or directory\n", targetPath)
		return false
	}

	// 检查是否是目录
	if !fileInfo.IsDir() {
		fmt.Fprintf(errorWriter, "cd: %s: No such file or directory\n", targetPath)
		return false
	}

	// 尝试切换目录
	err = os.Chdir(fullPath)
	if err != nil {
		fmt.Fprintf(errorWriter, "cd: %s: No such file or directory\n", targetPath)
		return false
	}

	return true
}

// 处理 exit 命令
func HandleExit() bool {
	SaveCmdHistoryToEnvFile()
	return true // 返回 true 表示应该退出循环
}

// 处理 echo 命令
func HandleEcho(realCommand string, outputFile *os.File, appendFile *os.File) bool {
	// 找到 "echo " 之后的内容（包括空格）
	echoPrefix := "echo "
	if !strings.HasPrefix(realCommand, echoPrefix) {
		return false
	}
	// 获取 echo 之后的参数部分
	words := realCommand[len(echoPrefix):]

	writer := io.Writer(os.Stdout)
	if outputFile != nil {
		writer = outputFile
	} else if appendFile != nil {
		writer = appendFile
	}

	runEchoBuiltin("echo "+words, writer)
	return true
}

// 处理 type 命令
func HandleType(cmdSlice []string, outputFile *os.File) bool {
	writer := os.Stdout
	if outputFile != nil {
		writer = outputFile
	}
	return runTypeBuiltin(cmdSlice, writer)
}

// 处理 pwd 命令
func HandlePwd(outputFile *os.File, errorFile *os.File, appendFile *os.File, errorAppendFile *os.File) bool {
	writer := os.Stdout
	if outputFile != nil {
		writer = outputFile
	} else if appendFile != nil {
		writer = appendFile
	}
	errorWriter := os.Stderr
	if errorFile != nil {
		errorWriter = errorFile
	} else if errorAppendFile != nil {
		errorWriter = errorAppendFile
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(errorWriter, "Error getting current directory: %v\n", err)
		return false
	}
	fmt.Fprintln(writer, wd)
	return true // 返回 true 表示应该 continue
}

func HandleHistory(actualCmdSlice []string, outputFile *os.File, appendFile *os.File) bool {
	writer := os.Stdout
	if outputFile != nil {
		writer = outputFile
	} else if appendFile != nil {
		writer = appendFile
	}
	errorWriter := os.Stderr

	if len(actualCmdSlice) == 1 {
		for i, cmd := range HistoryCmdSlice {
			fmt.Fprintf(writer, "%d  %s\n", i+1, cmd)
		}
	} else if len(actualCmdSlice) == 2 && actualCmdSlice[1] != "-r" {
		historyNum := actualCmdSlice[1]
		historyNumInt, err := strconv.Atoi(historyNum)
		if err != nil {
			fmt.Fprintf(errorWriter, "history: %s: invalid number\n", historyNum)
			return false
		}
		for i := len(HistoryCmdSlice) - historyNumInt; i < len(HistoryCmdSlice); i++ {
			fmt.Fprintf(writer, "%d  %s\n", i+1, HistoryCmdSlice[i])
		}
	} else if len(actualCmdSlice) >= 3 && actualCmdSlice[1] == "-r" {
		filePath := actualCmdSlice[2]
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(errorWriter, "history: failed to read %s: %v\n", filePath, err)
			return false
		}
		contents := strings.Split(string(data), "\n")
		for _, line := range contents {
			if strings.TrimSpace(line) == "" {
				continue
			}
			HistoryCmdSlice = append(HistoryCmdSlice, line)
		}
	} else if len(actualCmdSlice) >= 3 && actualCmdSlice[1] == "-w" {
		filePath := actualCmdSlice[2]
		contents := strings.Join(HistoryCmdSlice, "\n")
		contents += "\n" // 加上尾部换行符
		os.WriteFile(filePath, []byte(contents), 0644)
	} else if len(actualCmdSlice) >= 3 && actualCmdSlice[1] == "-a" {
		filePath := actualCmdSlice[2]
		appendFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(errorWriter, "history: failed to open %s: %v\n", filePath, err)
			return false
		}
		defer appendFile.Close()

		// 仅追加自上次写入之后的新历史记录，避免重复写入
		if lastHistoryWrittenIndex < 0 || lastHistoryWrittenIndex > len(HistoryCmdSlice) {
			lastHistoryWrittenIndex = 0
		}
		newEntries := HistoryCmdSlice[lastHistoryWrittenIndex:]
		if len(newEntries) > 0 {
			contents := strings.Join(newEntries, "\n") + "\n"
			if _, err = appendFile.Write([]byte(contents)); err != nil {
				fmt.Fprintf(errorWriter, "history: failed to write %s: %v\n", filePath, err)
				return false
			}
			lastHistoryWrittenIndex = len(HistoryCmdSlice)
		}
	} else {
		fmt.Fprintln(errorWriter, "history: invalid usage")
		return false
	}
	return true
}

// 处理外部命令
func HandleExternalCommand(commandName string, cmdSlice []string, outputFile *os.File,
	errorFile *os.File, appendFile *os.File, errorAppendFile *os.File) {
	if fullPath, found := utils.FindExecutable(commandName); found {
		// 执行外部程序，正确传递参数
		cmd := exec.Command(fullPath, cmdSlice[1:]...)
		// 设置标准输出：如果有重定向文件，使用文件；否则使用标准输出
		if outputFile != nil {
			cmd.Stdout = outputFile
		} else if appendFile != nil {
			cmd.Stdout = appendFile
		} else {
			cmd.Stdout = os.Stdout
		}
		// 设置标准错误：如果有重定向文件，使用文件；否则使用标准错误
		if errorFile != nil {
			cmd.Stderr = errorFile
		} else if errorAppendFile != nil {
			cmd.Stderr = errorAppendFile
		} else {
			cmd.Stderr = os.Stderr
		}

		// 设置进程属性，使外部程序看到的 argv[0] 是命令名而不是完整路径(fullPath)
		cmd.Args = append([]string{commandName}, cmdSlice[1:]...)

		// 执行命令，错误信息由命令本身输出到 stderr 或 errorFile
		// 不需要额外打印错误信息
		_ = cmd.Run()
	} else {
		fmt.Printf("%s: command not found\n", commandName)
	}
}
