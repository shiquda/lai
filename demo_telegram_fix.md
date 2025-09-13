# Telegram Entity Parsing Fix Demo

## 问题描述
修复了Telegram通知中的"Bad Request: can't parse entities: Can't find end of the entity starting at byte offset 83"错误。

## 根本原因
该错误是由于Telegram Markdown格式化逻辑中的以下问题导致的：

1. **不正确的字符转义**：特殊字符（`_`, `*`, `[`, `` ` ``）没有正确转义
2. **格式化实体不完整**：Markdown实体（如粗体、斜体）没有正确闭合
3. **链接处理问题**：链接格式的文本被不当地转义

## 修复方案

### 1. 重写了 `convertToTelegramMarkdown` 函数
- 改进了Markdown元素的处理顺序
- 确保所有格式化实体都正确闭合
- 更好地处理各种Markdown语法

### 2. 新增了专门的转义函数
- `escapeTelegramMarkdownContent`: 转义Markdown内容中的特殊字符
- `escapeTelegramCodeContent`: 处理代码块中的特殊字符

### 3. 改进了格式化检测逻辑
- 使用正则表达式准确检测Markdown格式
- 只对非Markdown文本进行转义
- 保持有效的Markdown格式不变

## 测试案例

### 输入示例
```
**Error**: Something went wrong with `test_file.log` at line 123. See [docs](https://example.com) for help.
```

### 修复前的问题
- 特殊字符会导致Telegram解析失败
- 格式化实体可能不完整
- 链接文本被错误转义

### 修复后的输出
```
*Error*: Something went wrong with `test_file.log` at line 123. See [docs](https://example.com) for help.
```

## 支持的Markdown格式
- **粗体**: `**text**` → `*text*`
- *斜体*: `_text_` → `_text_`
- # 标题: `# Heading` → `*Heading*`
- `代码`: `inline code` → `inline code`
- [链接](URL): `[text](url)` → `[text](url)`
- 列表: `- item` → `• item`

## 兼容性
- 保持与现有Telegram Legacy Markdown格式的兼容性
- 不破坏现有的通知功能
- 所有现有测试继续通过

## 测试验证
创建了全面的测试套件，验证了各种场景：
- 基本格式化（粗体、斜体）
- 标题转换
- 代码块处理
- 链接格式
- 特殊字符转义
- 混合内容处理

所有测试案例均通过验证。