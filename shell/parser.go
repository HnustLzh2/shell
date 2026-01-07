package shell

import "strings"

// ParsePipe 解析管道操作符，将命令分割成多个部分
// 返回命令部分列表，每个部分是一个命令及其参数
func ParsePipe(command string) []string {
	var parts []string
	var currentPart strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	bytes := []byte(command)

	for i := 0; i < len(bytes); i++ {
		if bytes[i] == '\'' {
			if inSingleQuotes {
				inSingleQuotes = false
				if i+1 < len(bytes) && bytes[i+1] == '\'' {
					i++
					inSingleQuotes = true
					continue
				}
			} else if !inDoubleQuotes {
				inSingleQuotes = true
			}
			currentPart.WriteByte(bytes[i])
		} else if bytes[i] == '"' {
			if inDoubleQuotes {
				inDoubleQuotes = false
				if i+1 < len(bytes) && bytes[i+1] == '"' {
					i++
					inDoubleQuotes = true
					continue
				}
			} else if !inSingleQuotes {
				inDoubleQuotes = true
			}
			currentPart.WriteByte(bytes[i])
		} else if !inSingleQuotes && !inDoubleQuotes && bytes[i] == '|' {
			// 遇到管道操作符，保存当前部分并开始新部分
			if currentPart.Len() > 0 {
				parts = append(parts, strings.TrimSpace(currentPart.String()))
				currentPart.Reset()
			}
		} else {
			currentPart.WriteByte(bytes[i])
		}
	}

	// 添加最后一部分
	if currentPart.Len() > 0 {
		parts = append(parts, strings.TrimSpace(currentPart.String()))
	}

	return parts
}

// ParseCommand 解析命令，处理单引号和双引号
func ParseCommand(command string) []string {
	var args []string
	var currentArg strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	bytes := []byte(command)

	for i := 0; i < len(bytes); i++ {
		if bytes[i] == '\'' {
			if inSingleQuotes {
				// 结束单引号
				inSingleQuotes = false
				// 检查下一个字符是否是引号（相邻引号字符串或空引号）
				if i+1 < len(bytes) && bytes[i+1] == '\'' {
					// 相邻引号字符串 'hello''world' 或空引号 ''
					i++ // 跳过下一个引号
					inSingleQuotes = true
					continue
				}
			} else if !inDoubleQuotes {
				// 开始单引号（不在双引号内）
				// 如果当前有参数，先保存
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
				// 检查是否是空引号 ''
				if i+1 < len(bytes) && bytes[i+1] == '\'' {
					// 空引号，忽略
					i++ // 跳过下一个引号
					continue
				}
				inSingleQuotes = true
			} else {
				// 在双引号内的单引号，按字面处理
				currentArg.WriteByte(bytes[i])
			}
		} else if bytes[i] == '"' {
			if inDoubleQuotes {
				// 结束双引号
				inDoubleQuotes = false
				// 检查下一个字符是否是引号（相邻引号字符串或空引号）
				if i+1 < len(bytes) && bytes[i+1] == '"' {
					// 相邻引号字符串 "hello""world" 或空引号 ""
					i++ // 跳过下一个引号
					inDoubleQuotes = true
					continue
				}
			} else if !inSingleQuotes {
				// 开始双引号（不在单引号内）
				// 如果当前有参数，先保存
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
				// 检查是否是空引号 ""
				if i+1 < len(bytes) && bytes[i+1] == '"' {
					// 空引号，忽略
					i++ // 跳过下一个引号
					continue
				}
				inDoubleQuotes = true
			} else {
				// 在单引号内的双引号，按字面处理
				currentArg.WriteByte(bytes[i])
			}
		} else if !inSingleQuotes && !inDoubleQuotes && (bytes[i] == ' ' || bytes[i] == '\t') {
			// 引号外的空格，作为分隔符
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		} else if inDoubleQuotes && bytes[i] == '\\' {
			// 在双引号内遇到反斜杠，处理转义
			if i+1 < len(bytes) {
				switch bytes[i+1] {
				case '"':
					currentArg.WriteByte('"')
					i++ // 跳过下一个字符
				case '\\':
					currentArg.WriteByte('\\')
					i++ // 跳过下一个字符
				default:
					currentArg.WriteByte(bytes[i])
				}
			} else {
				// 反斜杠是最后一个字符，按字面处理
				currentArg.WriteByte(bytes[i])
			}
		} else {
			// 普通字符，添加到当前参数
			currentArg.WriteByte(bytes[i])
		}
	}

	// 添加最后一个参数
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

// ParseRedirect 解析重定向操作符，返回命令部分、标准输出文件和标准错误文件
func ParseRedirect(cmdSlice []string) ([]string, string, string, string, string) {
	var stdoutFile string
	var stdappendFile string
	var stderrFile string
	var stderrappendFile string
	actualCmdSlice := cmdSlice

	// 从后往前查找重定向操作符（支持多个重定向）
	for i := len(cmdSlice) - 1; i >= 0; i-- {
		arg := cmdSlice[i]
		switch arg {
		case ">", "1>":
			if i+1 < len(cmdSlice) {
				stdoutFile = cmdSlice[i+1]
				actualCmdSlice = cmdSlice[:i]
			}
		case "2>":
			if i+1 < len(cmdSlice) {
				stderrFile = cmdSlice[i+1]
				actualCmdSlice = cmdSlice[:i]
			}
		case ">>", "1>>":
			if i+1 < len(cmdSlice) {
				stdappendFile = cmdSlice[i+1]
				actualCmdSlice = cmdSlice[:i]
			}
		case "2>>":
			if i+1 < len(cmdSlice) {
				stderrappendFile = cmdSlice[i+1]
				actualCmdSlice = cmdSlice[:i]
			}
		default:
			if strings.HasPrefix(arg, "1>>") {
				stdappendFile = strings.TrimPrefix(arg, "1>>")
				actualCmdSlice = append(cmdSlice[:i], cmdSlice[i+1:]...)
			} else if strings.HasPrefix(arg, ">>") {
				stdappendFile = strings.TrimPrefix(arg, ">>")
				actualCmdSlice = append(cmdSlice[:i], cmdSlice[i+1:]...)
			} else if strings.HasPrefix(arg, "1>") {
				stdoutFile = strings.TrimPrefix(arg, "1>")
				actualCmdSlice = append(cmdSlice[:i], cmdSlice[i+1:]...)
			} else if strings.HasPrefix(arg, ">") {
				stdoutFile = strings.TrimPrefix(arg, ">")
				actualCmdSlice = append(cmdSlice[:i], cmdSlice[i+1:]...)
			} else if strings.HasPrefix(arg, "2>>") {
				stderrappendFile = strings.TrimPrefix(arg, "2>>")
				actualCmdSlice = append(cmdSlice[:i], cmdSlice[i+1:]...)
			} else if strings.HasPrefix(arg, "2>") {
				stderrFile = strings.TrimPrefix(arg, "2>")
				actualCmdSlice = append(cmdSlice[:i], cmdSlice[i+1:]...)
			}
		}
	}

	return actualCmdSlice, stdoutFile, stderrFile, stdappendFile, stderrappendFile
}
