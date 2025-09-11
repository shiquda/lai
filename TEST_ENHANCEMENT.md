# Lai 测试功能增强文档

## 概述

Lai 的测试功能已经得到了显著增强，现在提供了更好的用户体验、详细的配置验证和多种测试模式。

## 新增功能

### 1. 多种测试模式

#### 标准测试模式
```bash
# 测试所有配置的通知器
lai test

# 测试特定的通知器
lai test --notifiers telegram,email

# 使用自定义消息
lai test --message "自定义测试消息"
```

#### 连接测试模式
```bash
# 仅测试连接性，不发送实际消息
lai test --connection-only
```

#### 详细诊断模式
```bash
# 显示详细的诊断信息和故障排除建议
lai test --diagnostic
```

#### 配置验证模式
```bash
# 仅验证配置，不进行实际测试
lai test --validate-only
```

### 2. 配置列表功能

```bash
# 显示所有可用的通知通道
lai test --list
```

示例输出：
```
📋 Available Notification Channels
=====================================
Team Chat:
  ❌ Disabled discord (Discord)
  ❌ Disabled slack (Slack)

Email:
  ❌ Disabled email (SMTP Email)

Messaging:
  ❌ Disabled telegram (Telegram Bot)
```

### 3. 详细配置信息

在详细模式或诊断模式下，系统会显示每个通知器的详细配置信息：

```
📋 Configuration Overview
=====================================
Test Mode: Diagnostic
Services to Test: 1
✅ telegram: Telegram Bot
   Status: Enabled
   Provider: telegram
   bot_token: glob****ken
   chat_id: -100global
```

### 4. 配置验证

系统会自动检查每个通知器的配置完整性：

```
❌ Configuration validation failed:
  - telegram: missing required configuration keys: bot_token
```

### 5. 智能故障排除

当测试失败时，系统会提供针对性的故障排除建议：

```
❌ telegram service test failed: failed to create telegram service: Not Found
   💡 Troubleshooting Tips:
   - Verify your API token/key is correct and not expired
   - Check if the token has proper permissions
```

## 使用示例

### 基本用法

```bash
# 1. 首先查看可用的通知器
lai test --list

# 2. 验证配置是否正确
lai test --validate-only

# 3. 测试连接性
lai test --connection-only

# 4. 完整测试
lai test --diagnostic
```

### 高级用法

```bash
# 测试特定通知器并显示详细信息
lai test --notifiers telegram,slack --diagnostic

# 使用自定义消息进行测试
lai test --message "🚨 紧急测试通知" --verbose

# 快速验证配置
lai test --validate-only --notifiers telegram
```

## 测试输出示例

### 成功示例
```
🧪 Lai Notification Test - Standard Mode
=====================================
📋 Configuration Overview
=====================================
Test Mode: Standard
Services to Test: 1
✅ telegram: Telegram Bot
   Status: Enabled
   Provider: telegram
   bot_token: 1234****5678
   chat_id: -100123456789

🔍 Testing telegram service...
✅ telegram service test succeeded (1.234s)

=====================================
📊 Test Summary
=====================================
Test Mode: Standard
Total Duration: 1.5s
Services Tested: 1
✅ Successful: 1
❌ Failed: 0
⚠️  Skipped: 0
```

### 失败示例
```
🧪 Lai Notification Test - Diagnostic Mode
=====================================
📋 Configuration Overview
=====================================
Test Mode: Diagnostic
Services to Test: 1
❌ telegram: Telegram Bot
   Status: Enabled
   Provider: telegram
   ⚠️  Missing required keys: bot_token

🔍 Testing telegram service...
❌ telegram service test failed
   Duration: 0.123s
   Error: telegram bot_token is required
   💡 Troubleshooting Tips:
   - Verify your API token/key is correct and not expired
   - Check if the token has proper permissions

=====================================
📊 Test Summary
=====================================
Test Mode: Diagnostic
Total Duration: 0.5s
Services Tested: 1
✅ Successful: 0
❌ Failed: 1
⚠️  Skipped: 0

❌ Failed Services:
  - telegram: Test failed after 123ms: telegram bot_token is required

💡 Troubleshooting Suggestions:
  - Use --diagnostic flag for detailed error information
  - Use --validate-only to check configuration
  - Check your configuration with 'lai config list'
  - Verify API keys and tokens are correct
```

## 支持的通知器

系统支持以下类型的通知器：

### 即时消息
- **Telegram Bot** - 需要 `bot_token` 和 `chat_id`

### 团队聊天
- **Slack** - 需要 `oauth_token`
- **Slack Webhook** - 需要 `webhook_url`
- **Discord** - 需要 `bot_token`
- **Discord Webhook** - 需要 `webhook_url`

### 邮件
- **SMTP Email** - 需要 `host`, `username`, `password`
- **Gmail** - 需要 `host`, `username`, `password`
- **SendGrid** - 需要 `api_key`, `from_email`
- **Mailgun** - 需要 `api_key`, `domain`

### 短信/推送
- **Pushover** - 需要 `token`, `user`
- **Twilio SMS** - 需要 `account_sid`, `auth_token`, `from_number`

### 监控服务
- **PagerDuty** - 需要 `routing_key`

### 企业通讯
- **DingTalk** - 需要 `access_token`
- **WeChat** - 需要 `corp_id`, `corp_secret`, `agent_id`

## 配置要求

每个通知器都有特定的配置要求。系统会自动验证以下内容：

1. **必需字段** - 检查所有必需的配置参数是否存在
2. **字段格式** - 验证参数格式是否正确
3. **服务状态** - 确认服务是否已启用

## 安全考虑

- **敏感信息屏蔽** - API密钥和令牌在输出中会被部分屏蔽
- **配置验证** - 在发送任何消息前验证配置
- **连接超时** - 所有网络请求都有30秒超时保护

## 故障排除

### 常见问题

1. **配置错误**
   ```bash
   # 验证配置
   lai test --validate-only
   ```

2. **连接问题**
   ```bash
   # 测试连接性
   lai test --connection-only
   ```

3. **权限问题**
   ```bash
   # 获取详细诊断
   lai test --diagnostic
   ```

### 错误代码

- **Missing required configuration keys** - 缺少必需的配置参数
- **Service is disabled** - 服务未启用
- **Not Found** - API密钥无效或服务不可用
- **timeout** - 请求超时
- **network** - 网络连接问题

## 最佳实践

1. **配置前先验证**
   ```bash
   lai test --validate-only
   ```

2. **使用诊断模式调试**
   ```bash
   lai test --diagnostic
   ```

3. **定期测试连接**
   ```bash
   lai test --connection-only
   ```

4. **监控多个通知器**
   ```bash
   lai test --notifiers telegram,slack,email
   ```

## 技术实现

### 新增的结构体

- `TestResult` - 单个服务测试结果
- `TestStatus` - 整体测试状态

### 新增的接口方法

- `IsServiceEnabled(serviceName string) bool` - 检查服务是否启用
- `GetServiceConfig(serviceName string) (map[string]interface{}, bool)` - 获取服务配置

### 支持的参数

- `--notifiers` - 指定测试的通知器
- `--message` - 自定义测试消息
- `--verbose` - 详细输出
- `--connection-only` - 仅测试连接
- `--diagnostic` - 诊断模式
- `--validate-only` - 仅验证配置
- `--list` - 列出可用通知器

## 兼容性

- 保持向后兼容性
- 现有配置文件仍然有效
- 原有的测试命令继续工作

## 更新日志

- ✅ 添加多种测试模式
- ✅ 增强配置验证
- ✅ 改进错误提示
- ✅ 添加故障排除建议
- ✅ 支持配置列表功能
- ✅ 敏感信息屏蔽
- ✅ 智能错误分析