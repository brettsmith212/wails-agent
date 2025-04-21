package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type RunShellCommandInput struct {
	Command string `json:"command" jsonschema_description:"The shell command to execute"`
	Cwd     string `json:"cwd,omitempty" jsonschema_description:"Optional working directory for the command. Defaults to current directory if not provided."`
}

func RunShellCommand(input json.RawMessage) (string, error) {
	shellCommandInput := RunShellCommandInput{}
	err := json.Unmarshal(input, &shellCommandInput)
	if err != nil {
		return "", err
	}

	if shellCommandInput.Command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	// Default to current directory if no working directory specified
	cwd := "."
	if shellCommandInput.Cwd != "" {
		cwd = shellCommandInput.Cwd
	}

	// Use bash to support pipes, redirects and other shell features
	cmd := exec.Command("bash", "-c", shellCommandInput.Command)
	cmd.Dir = cwd

	// Capture stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), fmt.Errorf("command execution failed: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

var RunShellCommandDefinition = ToolDefinition{
	Name: "run_shell_command",
	Description: `Execute a shell command and return its output.

This tool runs the provided command in a bash shell, allowing for pipes, redirects and other shell features.
The command is executed in the current directory unless a working directory is specified via the 'cwd' parameter.
Both stdout and stderr are captured and returned.

Use this tool when you need to interact with the system, such as compiling code, running tests, or checking system status.
`,
	InputSchema: GenerateSchema[RunShellCommandInput](),
	Function:    RunShellCommand,
}