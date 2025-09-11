package cmd

import (
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// ParseCommand parses a shell command string into command and arguments
// It properly handles quotes, escapes, and other shell syntax
func ParseCommand(commandStr string) (command string, args []string, err error) {
	// Parse the command using shell parser
	parser := syntax.NewParser()
	var words []*syntax.Word

	// Create a simple command statement
	stmt, err := parser.Parse(strings.NewReader(commandStr), "")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse command: %w", err)
	}

	// Extract command and arguments from the parsed statement
	if len(stmt.Stmts) > 0 {
		if cmd, ok := stmt.Stmts[0].Cmd.(*syntax.CallExpr); ok {
			for _, word := range cmd.Args {
				// Convert syntax.Word to string
				argStr, err := wordToString(word)
				if err != nil {
					return "", nil, fmt.Errorf("failed to convert argument: %w", err)
				}
				if argStr != "" {
					words = append(words, word)
				}
			}
		}
	}

	if len(words) == 0 {
		return "", nil, fmt.Errorf("no command found")
	}

	// Convert first word to command
	command, err = wordToString(words[0])
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert command: %w", err)
	}

	// Convert remaining words to arguments
	for i := 1; i < len(words); i++ {
		arg, err := wordToString(words[i])
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert argument %d: %w", i, err)
		}
		args = append(args, arg)
	}

	return command, args, nil
}

// wordToString converts a syntax.Word to a string
func wordToString(word *syntax.Word) (string, error) {
	if word == nil {
		return "", nil
	}

	var result strings.Builder
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			result.WriteString(p.Value)
		case *syntax.SglQuoted:
			result.WriteString(p.Value)
		case *syntax.DblQuoted:
			// Handle double quotes with potential expansions
			for _, quotedPart := range p.Parts {
				switch qp := quotedPart.(type) {
				case *syntax.Lit:
					result.WriteString(qp.Value)
				default:
					// Skip complex expansions for now
					return "", fmt.Errorf("complex quote expansions not supported")
				}
			}
		default:
			// Skip complex expansions for now
			return "", fmt.Errorf("complex word parts not supported")
		}
	}

	return result.String(), nil
}

// SimpleParseCommand provides a fallback simple parser for basic cases
// This handles the most common cases without full shell parsing overhead
func SimpleParseCommand(commandStr string) (command string, args []string, err error) {
	var words []string
	var currentWord strings.Builder
	var inSingleQuote bool
	var inDoubleQuote bool
	var escapeNext bool

	for _, r := range commandStr {
		if escapeNext {
			// Handle escaped characters
			if r == ' ' {
				// For escaped spaces, just add the space without breaking the word
				currentWord.WriteRune(r)
			} else if r == 'n' {
				currentWord.WriteRune('\n')
			} else if r == 't' {
				currentWord.WriteRune('\t')
			} else if r == 'r' {
				currentWord.WriteRune('\r')
			} else {
				// For other escaped characters, just add the character
				currentWord.WriteRune(r)
			}
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '\'':
			if inDoubleQuote {
				currentWord.WriteRune(r)
			} else {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if inSingleQuote {
				currentWord.WriteRune(r)
			} else {
				inDoubleQuote = !inDoubleQuote
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				currentWord.WriteRune(r)
			} else {
				if currentWord.Len() > 0 {
					words = append(words, currentWord.String())
					currentWord.Reset()
				}
			}
		default:
			currentWord.WriteRune(r)
		}
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	if escapeNext {
		return "", nil, fmt.Errorf("unterminated escape sequence")
	}
	if inSingleQuote {
		return "", nil, fmt.Errorf("unterminated single quote")
	}
	if inDoubleQuote {
		return "", nil, fmt.Errorf("unterminated double quote")
	}

	if len(words) == 0 {
		return "", nil, fmt.Errorf("no command found")
	}

	return words[0], words[1:], nil
}

// ParseCommandWrapper tries the full shell parser first, falls back to simple parser
func ParseCommandWrapper(commandStr string) (command string, args []string, err error) {
	// Try full shell parser first
	command, args, err = ParseCommand(commandStr)
	if err == nil {
		return command, args, nil
	}

	// Fall back to simple parser
	return SimpleParseCommand(commandStr)
}