# Configuration Reference

Complete guide to configuring Lai for your monitoring needs.

## Global Configuration File

Lai uses a global configuration file located at `~/.lai/config.yaml`. This file is automatically created when you first run the configuration commands.

### Configuration Structure

```yaml
notifications:
  openai:
    api_key: "your-openai-api-key"
    base_url: "https://api.openai.com/v1"  # Optional
    model: "gpt-4o"                        # Optional
  telegram:
    bot_token: "your-telegram-bot-token"
    chat_id: "your-default-chat-id"

defaults:
  line_threshold: 10                       # Number of new lines to trigger summary
  check_interval: "30s"                    # How often to check for changes
  chat_id: "your-default-chat-id"          # Default Telegram chat
  final_summary: true                      # Send summary when programs exit
```

## Configuration Management

### Using CLI Commands

```bash
# Set OpenAI API key
lai config set notifications.openai.api_key "sk-your-key"

# Set OpenAI base URL (for custom endpoints)
lai config set notifications.openai.base_url "https://api.openai.com/v1"

# Set OpenAI model
lai config set notifications.openai.model "gpt-4o"

# Set Telegram bot token
lai config set notifications.telegram.bot_token "123456:ABC-DEF"

# Set default chat ID
lai config set defaults.chat_id "-100123456789"

# Set default line threshold
lai config set defaults.line_threshold 20

# Set default check interval
lai config set defaults.check_interval "60s"

# Enable/disable final summary by default
lai config set defaults.final_summary true

# View current configuration
lai config list

# Reset configuration to defaults
lai config reset
```

### Manual Configuration

You can also edit the configuration file directly at `~/.lai/config.yaml`.

## Configuration Options

### OpenAI Settings

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `api_key` | Your OpenAI API key | - | ✅ |
| `base_url` | API endpoint URL | `https://api.openai.com/v1` | ❌ |
| `model` | GPT model to use | `gpt-4o` | ❌ |

### Telegram Settings

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `bot_token` | Telegram bot token | - | ✅ |
| `chat_id` | Default chat/group ID | - | ✅ |

### Default Settings

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `line_threshold` | Lines needed to trigger summary | `10` | ❌ |
| `check_interval` | Check frequency | `30s` | ❌ |
| `chat_id` | Default Telegram chat | - | ❌ |
| `final_summary` | Send summary on program exit | `true` | ❌ |

## Setup Guides

### Getting OpenAI API Key

1. Visit [OpenAI Platform](https://platform.openai.com/)
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy and set it: `lai config set notifications.openai.api_key "sk-your-key"`

### Setting up Telegram Bot

1. **Create a Bot**:
   - Message [@BotFather](https://t.me/botfather) on Telegram
   - Send `/newbot`
   - Follow instructions to create your bot
   - Save the bot token

2. **Get Chat ID**:
   - For personal chats: Message your bot, then visit `https://api.telegram.org/bot<BOT_TOKEN>/getUpdates`
   - For groups: Add bot to group, send a message, then check the same URL
   - Look for the `chat.id` field

3. **Configure Lai**:
   ```bash
   lai config set notifications.telegram.bot_token "123456:ABC-DEF"
   lai config set defaults.chat_id "-100123456789"
   ```

### Using Custom OpenAI Endpoints

For custom OpenAI-compatible endpoints (like Azure OpenAI, LocalAI, etc.):

```bash
# Set custom base URL
lai config set notifications.openai.base_url "https://your-endpoint.com/v1"

# Set appropriate model name
lai config set notifications.openai.model "your-model"
```

## Command-Line Overrides

Most settings can be overridden per command:

```bash
# Override line threshold and interval
lai start /path/to/log --line-threshold 5 --interval 15s

# Override chat ID
lai start /path/to/log --chat-id "-100987654321"

# Override final summary setting
lai exec "npm test" --final-summary
lai exec "npm test" --no-final-summary
```

## Environment Variables

Sensitive configuration can also be set via environment variables:

```bash
export LAI_OPENAI_API_KEY="sk-your-key"
export LAI_TELEGRAM_BOT_TOKEN="123456:ABC-DEF"
export LAI_DEFAULT_CHAT_ID="-100123456789"

# Then run without explicit config
lai start /path/to/log
```

## Validation

Lai validates your configuration on startup. Common validation errors:

- **Missing API Key**: `notifications.openai.api_key is required`
- **Missing Bot Token**: `notifications.telegram.bot_token is required`
- **Invalid Interval**: `check_interval must be a valid duration (e.g., "30s", "2m")`
- **Invalid Threshold**: `line_threshold must be a positive integer`

## Best Practices

1. **Security**: Never commit API keys to version control
2. **Performance**: Use appropriate intervals (30s-2m) to balance responsiveness and API costs
3. **Thresholds**: Start with 10 lines threshold and adjust based on your log volume
4. **Testing**: Use a test chat/channel when setting up to avoid spam