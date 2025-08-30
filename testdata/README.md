# Test configuration files

## config.yaml - Valid configuration
# Windows example: log_file: "C:\\temp\\test.log"  
# Unix example: log_file: "/tmp/test.log"
log_file: "/tmp/test.log"
line_threshold: 5
check_interval: "10s"

openai:
  api_key: "sk-test-key-123"
  base_url: "https://api.openai.com/v1" 
  model: "gpt-4o"

telegram:
  bot_token: "123456:ABC-DEF1234ghIKl-zyx57W2v1u123ew11"
  chat_id: "-100123456789"

## invalid_config.yaml - Missing required fields
line_threshold: 5
check_interval: "10s"

## global_config.yaml - Global configuration example
notifications:
  openai:
    api_key: "sk-global-key-456"
    base_url: "https://api.openai.com/v1"
    model: "gpt-3.5-turbo"
  telegram:
    bot_token: "987654:XYZ-ABC9876fed54lIJk-abc21v9u876fd32"
    chat_id: "-100987654321"

defaults:
  line_threshold: 15
  check_interval: "60s"
  chat_id: "-100987654321"