package shell

import (
	"fmt"
	"io"
	"strings"

	"go_shell/utils"
)

func isBuiltinCommand(name string) bool {
	for _, cmd := range ShellSlice {
		if cmd == name {
			return true
		}
	}
	return false
}
func runBuiltinCommand(cmd pipelineCommand, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	switch cmd.args[0] {
	case "echo":
		runEchoBuiltin(cmd.raw, stdout)
	case "type":
		runTypeBuiltin(cmd.args, stdout)
	default:
		fmt.Fprintf(stderr, "%s: unsupported pipeline builtin\n", cmd.args[0])
	}
	return nil
}

func runEchoBuiltin(realCommand string, writer io.Writer) {
	echoPrefix := "echo "
	if !strings.HasPrefix(realCommand, echoPrefix) {
		return
	}
	words := realCommand[len(echoPrefix):]

	var result strings.Builder
	bytes := []byte(words)
	inQuotes := false
	inDoubleQuotes := false
	lastWasSpace := false

	for i := 0; i < len(bytes); i++ {
		if bytes[i] == '\'' {
			if inQuotes {
				inQuotes = false
				lastWasSpace = false
				for i+1 < len(bytes) && bytes[i+1] == '\'' {
					i++
				}
			} else if !inDoubleQuotes {
				inQuotes = true
				lastWasSpace = false
			} else {
				result.WriteByte(bytes[i])
			}
		} else if inQuotes {
			result.WriteByte(bytes[i])
		} else if bytes[i] == '"' {
			if inDoubleQuotes {
				inDoubleQuotes = false
				lastWasSpace = false
				for i+1 < len(bytes) && bytes[i+1] == '"' {
					i++
				}
			} else if !inQuotes {
				inDoubleQuotes = true
				lastWasSpace = false
			} else {
				result.WriteByte(bytes[i])
			}
		} else if inDoubleQuotes {
			if bytes[i] == '\\' {
				if i+1 < len(bytes) {
					switch bytes[i+1] {
					case '"':
						result.WriteByte('"')
						i++
					case '\\':
						result.WriteByte('\\')
						i++
					default:
						result.WriteByte(bytes[i])
						i++
					}
				}
			} else {
				result.WriteByte(bytes[i])
			}
		} else if bytes[i] == '\\' && i+1 < len(bytes) && !inDoubleQuotes && !inQuotes {
			lastWasSpace = false
			switch bytes[i+1] {
			case 'n':
				result.WriteByte('n')
				i++
			case 't':
				result.WriteByte('t')
				i++
			case '\'':
				result.WriteByte('\'')
				i++
			case '"':
				result.WriteByte('"')
				i++
			case '\\':
				result.WriteByte('\\')
				i++
			case ' ':
				result.WriteByte(' ')
				i++
			default:
				result.WriteByte(bytes[i])
			}
		} else if bytes[i] == ' ' || bytes[i] == '\t' {
			if !lastWasSpace {
				result.WriteByte(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteByte(bytes[i])
			lastWasSpace = false
		}
	}

	output := result.String()
	fmt.Fprint(writer, output)
	if !strings.HasSuffix(output, "\n") {
		fmt.Fprintln(writer)
	}
}
func runTypeBuiltin(cmdSlice []string, writer io.Writer) bool {
	if len(cmdSlice) != 2 {
		return false
	}
	testedType := cmdSlice[1]
	isBuiltin := false
	for _, shell := range ShellSlice {
		if testedType == shell {
			isBuiltin = true
			fmt.Fprintf(writer, "%s is a shell builtin\n", shell)
			break
		}
	}
	if !isBuiltin {
		if fullPath, found := utils.FindExecutable(testedType); found {
			fmt.Fprintf(writer, "%s is %s\n", testedType, fullPath)
		} else {
			fmt.Fprintf(writer, "%s: not found\n", testedType)
		}
	}
	return true
}
