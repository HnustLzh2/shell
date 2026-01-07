package shell

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"go_shell/utils"
)

// CreateCompleter 创建自动补全器
func CreateCompleter(trie *utils.Trie) readline.AutoCompleter {
	return &CustomCompleter{trie: trie}
}

// CustomCompleter 自定义补全器
type CustomCompleter struct {
	trie       *utils.Trie
	lastPrefix string // 上一次的输入前缀
	tabPressed bool   // 是否已经按过一次 TAB（针对当前前缀）
}

func (c *CustomCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line)
	parts := strings.Fields(lineStr)
	if len(parts) == 0 {
		// 重置状态
		c.lastPrefix = ""
		c.tabPressed = false
		return nil, 0
	}
	prefix := parts[0]
	// 将前缀转换为小写进行查找（保持与存储格式一致）
	lowerPrefix := strings.ToLower(prefix)

	// 如果前缀改变了，重置 TAB 状态
	if lowerPrefix != c.lastPrefix {
		c.lastPrefix = lowerPrefix
		c.tabPressed = false
	}

	completions := c.trie.FindCompletions(lowerPrefix)
	if len(completions) == 1 {
		// 重置状态（因为找到了唯一匹配）
		c.tabPressed = false
		completion := completions[0]
		if completion != "" && completion != lowerPrefix {
			// 找到有效的补全，返回需要追加的部分
			remaining := completion[len(lowerPrefix):] + " "
			return [][]rune{[]rune(remaining)}, len(remaining)
		}
		if completion == lowerPrefix {
			// 命令已经完整（用户输入的就是完整的有效命令），不需要补全，也不输出铃声
			return nil, 0
		}
	} else if len(completions) > 1 {
		// 多个匹配：计算最长公共前缀
		commonPrefix := longestCommonPrefix(completions)

		// 如果最长公共前缀比用户输入的前缀长，则补全到最长公共前缀
		if len(commonPrefix) > len(lowerPrefix) {
			// 补全到最长公共前缀
			remaining := commonPrefix[len(lowerPrefix):]
			// 补全后仍有多个匹配，只补全到最长公共前缀
			c.tabPressed = false
			return [][]rune{[]rune(remaining)}, len(remaining)
		}

		// 最长公共前缀等于用户输入的前缀，无法进一步补全
		// 按原来的逻辑处理：第一次响铃，第二次显示列表
		sortedCompletions := make([]string, len(completions))
		copy(sortedCompletions, completions)
		sort.Strings(sortedCompletions)
		if !c.tabPressed {
			// 第一次按 TAB：响铃
			c.tabPressed = true
			fmt.Fprint(os.Stdout, "\x07")
		} else {
			// 第二次按 TAB：打印所有匹配的可执行文件
			// 用2个空格分隔，打印在下一行
			listStr := strings.Join(sortedCompletions, "  ")
			fmt.Fprintf(os.Stdout, "\n%s\n", listStr)
			// 在新行显示提示符和用户输入的内容
			fmt.Fprintf(os.Stdout, "$ %s", lineStr)
			// 重置状态
			c.tabPressed = false
		}
	} else {
		// 没有匹配：响铃
		c.tabPressed = false
		fmt.Fprint(os.Stdout, "\x07")
	}
	return nil, 0
}

// longestCommonPrefix 计算字符串数组的最长公共前缀
func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	// 找到最短字符串的长度
	minLen := len(strs[0])
	for _, s := range strs {
		if len(s) < minLen {
			minLen = len(s)
		}
	}

	// 逐个字符比较
	for i := 0; i < minLen; i++ {
		ch := strs[0][i]
		for j := 1; j < len(strs); j++ {
			if strs[j][i] != ch {
				return strs[0][:i]
			}
		}
	}

	// 所有字符串的前 minLen 个字符都相同
	return strs[0][:minLen]
}
