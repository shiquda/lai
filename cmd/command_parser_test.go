package cmd

import (
	"testing"
)

func TestParseCommandWrapper(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCmd    string
		expectedArgs   []string
		expectError    bool
		errorContains  string
	}{
		{
			name:         "Simple command",
			input:        "ls -la",
			expectedCmd:  "ls",
			expectedArgs: []string{"-la"},
			expectError:  false,
		},
		{
			name:         "Command with quoted file path",
			input:        "uv run src/pdf_to_epub.py 'data/Neuroscience_Lecture_Wang_Liming_A5.pdf'",
			expectedCmd:  "uv",
			expectedArgs: []string{"run", "src/pdf_to_epub.py", "data/Neuroscience_Lecture_Wang_Liming_A5.pdf"},
			expectError:  false,
		},
		{
			name:         "Command with double quoted path",
			input:        `python script.py "data/file with spaces.txt"`,
			expectedCmd:  "python",
			expectedArgs: []string{"script.py", "data/file with spaces.txt"},
			expectError:  false,
		},
		{
			name:         "Command with mixed quotes",
			input:        `echo 'hello world' "goodbye moon"`,
			expectedCmd:  "echo",
			expectedArgs: []string{"hello world", "goodbye moon"},
			expectError:  false,
		},
		{
			name:         "Command with escaped quotes",
			input:        `echo "hello \"world\""`,
			expectedCmd:  "echo",
			expectedArgs: []string{`hello \"world\"`},
			expectError:  false,
		},
		{
			name:         "Command with single quotes inside double quotes",
			input:        `echo "hello 'world'"`,
			expectedCmd:  "echo",
			expectedArgs: []string{`hello 'world'`},
			expectError:  false,
		},
		{
			name:         "Command with special characters",
			input:        `python -c "print('hello world')"`,
			expectedCmd:  "python",
			expectedArgs: []string{"-c", "print('hello world')"},
			expectError:  false,
		},
		{
			name:          "Unterminated single quote",
			input:         "echo 'hello world",
			expectError:   true,
			errorContains: "unterminated single quote",
		},
		{
			name:          "Unterminated double quote",
			input:         `echo "hello world`,
			expectError:   true,
			errorContains: "unterminated double quote",
		},
		{
			name:         "Unterminated escape",
			input:        "echo hello\\",
			expectedCmd:  "echo",
			expectedArgs: []string{"hello\\"},
			expectError:  false,
		},
		{
			name:          "Empty command",
			input:         "",
			expectError:   true,
			errorContains: "no command found",
		},
		{
			name:          "Only whitespace",
			input:         "   ",
			expectError:   true,
			errorContains: "no command found",
		},
		{
			name:         "Complex command with multiple quoted args",
			input:        `docker run -v "/host/path:/container/path" --name "my container" image command`,
			expectedCmd:  "docker",
			expectedArgs: []string{"run", "-v", "/host/path:/container/path", "--name", "my container", "image", "command"},
			expectError:  false,
		},
		{
			name:         "Command with escaped spaces",
			input:        `echo hello\ world`,
			expectedCmd:  "echo",
			expectedArgs: []string{`hello\ world`},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := ParseCommandWrapper(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsSubstring(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if cmd != tt.expectedCmd {
				t.Errorf("Expected command '%s' but got '%s'", tt.expectedCmd, cmd)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args but got %d", len(tt.expectedArgs), len(args))
				return
			}

			for i, expectedArg := range tt.expectedArgs {
				if args[i] != expectedArg {
					t.Errorf("Expected arg[%d] '%s' but got '%s'", i, expectedArg, args[i])
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		   findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}