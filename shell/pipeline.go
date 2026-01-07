package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"go_shell/utils"
)

type pipelineCommand struct {
	args      []string
	raw       string
	isBuiltin bool
}

type pipelineProcess interface {
	Start() error
	Wait() error
}

type externalProcess struct {
	cmd *exec.Cmd
}

func (p *externalProcess) Start() error {
	return p.cmd.Start()
}

func (p *externalProcess) Wait() error {
	return p.cmd.Wait()
}

type builtinProcess struct {
	cmd          pipelineCommand
	stdin        io.Reader
	stdout       io.Writer
	stdoutCloser io.Closer
	stderr       io.Writer
	done         chan error
}

func newBuiltinProcess(cmd pipelineCommand, stdin io.Reader, stdout io.Writer, stdoutCloser io.Closer, stderr io.Writer) *builtinProcess {
	return &builtinProcess{
		cmd:          cmd,
		stdin:        stdin,
		stdout:       stdout,
		stdoutCloser: stdoutCloser,
		stderr:       stderr,
	}
}

func (p *builtinProcess) Start() error {
	p.done = make(chan error, 1)
	go func() {
		err := runBuiltinCommand(p.cmd, p.stdin, p.stdout, p.stderr)
		if p.stdoutCloser != nil {
			p.stdoutCloser.Close()
		}
		p.done <- err
	}()
	return nil
}

func (p *builtinProcess) Wait() error {
	if p.done == nil {
		return nil
	}
	return <-p.done
}

// HandlePipeline 处理管道命令
func HandlePipeline(pipeParts []string) {
	if len(pipeParts) < 2 {
		return
	}

	commands := make([]pipelineCommand, len(pipeParts))

	var stdoutFile, stderrFile, stdappendFile, stderrappendFile string

	for i, part := range pipeParts {
		cmdSlice := ParseCommand(part)
		if len(cmdSlice) == 0 {
			return
		}

		actualCmdSlice, outFile, errFile, appendFile, errAppendFile := ParseRedirect(cmdSlice)
		if len(actualCmdSlice) == 0 {
			return
		}

		commands[i] = pipelineCommand{
			args:      actualCmdSlice,
			raw:       strings.TrimSpace(part),
			isBuiltin: isBuiltinCommand(actualCmdSlice[0]),
		}

		if i == len(pipeParts)-1 {
			stdoutFile = outFile
			stderrFile = errFile
			stdappendFile = appendFile
			stderrappendFile = errAppendFile
		}
	}

	// 配置最终输出
	finalStdout := io.Writer(os.Stdout)
	var finalStdoutCloser io.Closer
	if stdoutFile != "" {
		outputFile, err := os.Create(stdoutFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file %s: %v\n", stdoutFile, err)
			return
		}
		finalStdout = outputFile
		finalStdoutCloser = outputFile
	} else if stdappendFile != "" {
		appendFile, err := os.OpenFile(stdappendFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", stdappendFile, err)
			return
		}
		finalStdout = appendFile
		finalStdoutCloser = appendFile
	}

	finalStderr := io.Writer(os.Stderr)
	var finalStderrCloser io.Closer
	if stderrFile != "" {
		errorFile, err := os.Create(stderrFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file %s: %v\n", stderrFile, err)
			if finalStdoutCloser != nil {
				finalStdoutCloser.Close()
			}
			return
		}
		finalStderr = errorFile
		finalStderrCloser = errorFile
	} else if stderrappendFile != "" {
		errAppendFile, err := os.OpenFile(stderrappendFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", stderrappendFile, err)
			if finalStdoutCloser != nil {
				finalStdoutCloser.Close()
			}
			return
		}
		finalStderr = errAppendFile
		finalStderrCloser = errAppendFile
	}

	executePipeline(commands, finalStdout, finalStderr)

	if finalStdoutCloser != nil {
		finalStdoutCloser.Close()
	}
	if finalStderrCloser != nil {
		finalStderrCloser.Close()
	}
}

// executePipeline 执行管道命令
func executePipeline(commands []pipelineCommand, finalStdout io.Writer, finalStderr io.Writer) {
	if len(commands) < 2 {
		return
	}

	// 创建管道
	pipeReaders := make([]*os.File, len(commands)-1)
	pipeWriters := make([]*os.File, len(commands)-1)

	for i := 0; i < len(commands)-1; i++ {
		reader, writer, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating pipe: %v\n", err)
			return
		}
		pipeReaders[i] = reader
		pipeWriters[i] = writer
	}

	processes := make([]pipelineProcess, len(commands))

	for i, cmdInfo := range commands {
		var stdin io.Reader
		if i == 0 {
			stdin = os.Stdin
		} else {
			stdin = pipeReaders[i-1]
		}

		var stdout io.Writer
		var stdoutCloser io.Closer
		if i == len(commands)-1 {
			stdout = finalStdout
		} else {
			stdout = pipeWriters[i]
			stdoutCloser = pipeWriters[i]
		}

		var stderr io.Writer
		if i == len(commands)-1 {
			stderr = finalStderr
		} else {
			stderr = os.Stderr
		}

		if cmdInfo.isBuiltin && (cmdInfo.args[0] == "echo" || cmdInfo.args[0] == "type") {
			processes[i] = newBuiltinProcess(cmdInfo, stdin, stdout, stdoutCloser, stderr)
		} else {
			fullPath, found := utils.FindExecutable(cmdInfo.args[0])
			if !found {
				fmt.Printf("%s: command not found\n", cmdInfo.args[0])
				for j := 0; j < len(pipeReaders); j++ {
					pipeReaders[j].Close()
				}
				for j := 0; j < len(pipeWriters); j++ {
					pipeWriters[j].Close()
				}
				return
			}
			cmd := exec.Command(fullPath, cmdInfo.args[1:]...)
			cmd.Args = append([]string{cmdInfo.args[0]}, cmdInfo.args[1:]...)
			cmd.Stdin = stdin
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			processes[i] = &externalProcess{cmd: cmd}
		}
	}

	for _, proc := range processes {
		if err := proc.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting command: %v\n", err)
			for j := range pipeReaders {
				pipeReaders[j].Close()
			}
			for j := range pipeWriters {
				pipeWriters[j].Close()
			}
			return
		}
	}

	for i := range pipeWriters {
		pipeWriters[i].Close()
	}

	for _, proc := range processes {
		proc.Wait()
	}

	for i := 0; i < len(pipeReaders); i++ {
		pipeReaders[i].Close()
	}
}
