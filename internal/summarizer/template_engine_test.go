package summarizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TemplateEngineTestSuite struct {
	suite.Suite
	engine *TemplateEngine
}

func (s *TemplateEngineTestSuite) SetupTest() {
	s.engine = NewTemplateEngine()
	s.engine.SetBuiltinVariable("language", "English")
	s.engine.SetBuiltinVariable("system", "Lai")
}

func TestTemplateEngineSuite(t *testing.T) {
	suite.Run(t, new(TemplateEngineTestSuite))
}

func (s *TemplateEngineTestSuite) TestNewTemplateEngine() {
	engine := NewTemplateEngine()
	assert.NotNil(s.T(), engine)
	assert.NotNil(s.T(), engine.builtinVariables)
	assert.Contains(s.T(), engine.builtinVariables, "language")
	assert.Contains(s.T(), engine.builtinVariables, "system")
}

func (s *TemplateEngineTestSuite) TestSetBuiltinVariable() {
	s.engine.SetBuiltinVariable("test_var", "test_value")
	assert.Equal(s.T(), "test_value", s.engine.builtinVariables["test_var"])
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_DoubleBraceFormat() {
	template := "Hello {{language}}, this is a test with {{log_content}}"
	variables := map[string]string{
		"log_content": "sample log data",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello English, this is a test with sample log data", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_DollarBraceFormat() {
	template := "Hello ${language}, log: ${log_content}"
	variables := map[string]string{
		"log_content": "error message",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello English, log: error message", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_SimpleDollarFormat() {
	template := "Hello $language, log: $log_content"
	variables := map[string]string{
		"log_content": "warning message",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello English, log: warning message", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_MixedFormats() {
	template := "Hello {{language}}, log: ${log_content}, system: $system"
	variables := map[string]string{
		"log_content": "mixed format test",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello English, log: mixed format test, system: Lai", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_CustomVariablesOverrideBuiltin() {
	template := "Hello {{language}}, from {{custom_source}}"
	variables := map[string]string{
		"language":      "Chinese", // Override built-in
		"custom_source": "Custom App",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello Chinese, from Custom App", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_UndefinedVariable() {
	template := "Hello {{language}}, from {{unknown_var}}"
	variables := map[string]string{}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	// Undefined variables should be left as-is for backward compatibility
	assert.Equal(s.T(), "Hello English, from {{unknown_var}}", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_EmptyTemplate() {
	_, err := s.engine.RenderTemplate("", map[string]string{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "template cannot be empty")
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_NoVariables() {
	template := "Simple text with no variables"
	result, err := s.engine.RenderTemplate(template, map[string]string{})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), template, result)
}

func (s *TemplateEngineTestSuite) TestValidateTemplate_ValidTemplate() {
	template := "Hello {{language}}, log: {{log_content}}"
	allowedVariables := map[string]bool{
		"language":    true,
		"log_content": true,
	}

	err := s.engine.ValidateTemplate(template, allowedVariables)
	assert.NoError(s.T(), err)
}

func (s *TemplateEngineTestSuite) TestValidateTemplate_UndefinedVariables() {
	template := "Hello {{language}}, log: {{unknown_var}}, system: $system"
	allowedVariables := map[string]bool{
		"language": true,
	}

	err := s.engine.ValidateTemplate(template, allowedVariables)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "undefined variables found")
	assert.Contains(s.T(), err.Error(), "unknown_var")
}

func (s *TemplateEngineTestSuite) TestValidateTemplate_BuiltinVariablesAllowed() {
	template := "Hello {{language}}, system: $system"
	allowedVariables := map[string]bool{
		"log_content": true,
	}

	err := s.engine.ValidateTemplate(template, allowedVariables)
	assert.NoError(s.T(), err) // Built-in variables are allowed
}

func (s *TemplateEngineTestSuite) TestValidateTemplate_EmptyTemplate() {
	err := s.engine.ValidateTemplate("", map[string]bool{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "template cannot be empty")
}

func (s *TemplateEngineTestSuite) TestGetBuiltinVariables() {
	variables := s.engine.GetBuiltinVariables()
	assert.NotNil(s.T(), variables)
	assert.Contains(s.T(), variables, "language")
	assert.Contains(s.T(), variables, "system")
}

func (s *TemplateEngineTestSuite) TestGetSupportedVariableFormats() {
	formats := s.engine.GetSupportedVariableFormats()
	expected := []string{"{{variable}}", "${variable}", "$variable"}
	assert.Equal(s.T(), expected, formats)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_WhitespaceHandling() {
	template := "Hello {{  language  }}, log: {{ log_content  }}"
	variables := map[string]string{
		"log_content": "whitespace test",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello English, log: whitespace test", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_SpecialCharactersInVariables() {
	template := "Test: {{special_var}}, More: {{another_var}}"
	variables := map[string]string{
		"special_var": "value with spaces",
		"another_var": "value-with-dashes_and_underscores",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Test: value with spaces, More: value-with-dashes_and_underscores", result)
}

func (s *TemplateEngineTestSuite) TestRenderTemplate_VariableNamesWithNumbers() {
	template := "Test: {{var1}}, More: {{var_2}}"
	variables := map[string]string{
		"var1":  "first value",
		"var_2": "second value",
	}

	result, err := s.engine.RenderTemplate(template, variables)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Test: first value, More: second value", result)
}
