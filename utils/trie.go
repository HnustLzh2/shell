package utils

import (
	"os"
	"strings"
)

type Trie struct {
	children map[rune]*Trie
	isEnd    bool
}

func Constructor() *Trie {
	return &Trie{
		children: make(map[rune]*Trie),
	}
}

func (t *Trie) Insert(word string) {
	node := t
	for _, ch := range word {
		if node.children == nil {
			node.children = make(map[rune]*Trie)
		}
		if node.children[ch] == nil {
			node.children[ch] = Constructor()
		}
		node = node.children[ch]
	}
	node.isEnd = true
}

func (t *Trie) SearchWithPrefix(prefix string) *Trie {
	node := t
	for _, ch := range prefix {
		if node.children == nil || node.children[ch] == nil {
			return nil
		}
		node = node.children[ch]
	}
	return node
}

func (t *Trie) Search(word string) bool {
	node := t.SearchWithPrefix(word)
	return node != nil && node.isEnd
}

func (t *Trie) StartsWith(prefix string) bool {
	return t.SearchWithPrefix(prefix) != nil
}

func (t *Trie) FindCompletions(prefix string) []string {
	node := t.SearchWithPrefix(prefix)
	if node == nil {
		return nil
	}
	var completions []string
	t.collectWords(node, prefix, &completions)
	return completions
}

// collectWords 收集从给定节点开始的所有完整单词
func (t *Trie) collectWords(node *Trie, prefix string, completions *[]string) {
	if node == nil {
		return
	}

	// 如果当前节点是完整单词，添加到结果中
	// 注意：即使当前节点是完整单词，也要继续查找子节点
	// 因为可能存在一个单词是另一个单词的前缀（如 xyz_fox 和 xyz_fox_rat）
	if node.isEnd {
		*completions = append(*completions, prefix)
	}

	// 遍历所有子节点（使用 map 后需要遍历 map）
	if node.children != nil {
		for ch, child := range node.children {
			if child != nil {
				t.collectWords(child, prefix+string(ch), completions)
			}
		}
	}
}

// 查找可执行文件的函数
func FindExecutable(command string) (string, bool) {
	pathEnv := os.Getenv("PATH")
	dirs := strings.Split(pathEnv, ":")
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		fullPath := dir + "/" + command
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		// 检查文件是否是常规文件并且有执行权限
		if fileInfo.Mode().IsRegular() && fileInfo.Mode()&0111 != 0 {
			return fullPath, true
		}
	}
	return "", false
}
