package cmd

import (
	"testing"
)

// BenchmarkParseCommand benchmarks the different parsing approaches
func BenchmarkParseCommand(b *testing.B) {
	testCases := []string{
		"ls -la",
		`uv run src/pdf_to_epub.py 'data/Neuroscience_Lecture_Wang_Liming_A5.pdf'`,
		`python script.py "data/file with spaces.txt"`,
		`echo 'hello world' "goodbye moon"`,
		`docker run -v "/host/path:/container/path" --name "my container" image command`,
		`echo "He said 'hello world'"`,
		`command "arg 1" 'arg 2' arg3`,
	}

	b.Run("ParseCommand", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				ParseCommand(tc)
			}
		}
	})

	b.Run("SimpleParseCommand", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				SimpleParseCommand(tc)
			}
		}
	})

	b.Run("ParseCommandWrapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, tc := range testCases {
				ParseCommandWrapper(tc)
			}
		}
	})
}

// TestParseCommandsComparison compares our implementation with potential alternatives
func TestParseCommandsComparison(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple command",
			input:    "ls -la",
			expected: []string{"ls", "-la"},
		},
		{
			name:     "Empty quoted args",
			input:    `echo "" ''`,
			expected: []string{"echo", "", ""},
		},
		{
			name:     "Complex quoting",
			input:    `echo 'hello "world"' "foo 'bar'"`,
			expected: []string{"echo", `hello "world"`, `foo 'bar'`},
		},
		{
			name:     "Concatenated quotes",
			input:    `echo "hello"'world'`,
			expected: []string{"echo", `helloworld`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, args, err := ParseCommandWrapper(tc.input)
			if err != nil {
				t.Fatalf("ParseCommandWrapper failed: %v", err)
			}

			result := append([]string{cmd}, args...)
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d parts, got %d", len(tc.expected), len(result))
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Part %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}