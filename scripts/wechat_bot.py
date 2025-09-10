#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
企业微信通知脚本
用于 Lai 日志监控的通知发送
"""

import requests
import json
import sys
import os
from datetime import datetime

class WeChatBot:
    def __init__(self, webhook_url):
        self.webhook_url = webhook_url
    
    def send_message(self, message, title="Lai 通知"):
        """发送消息到企业微信"""
        payload = {
            "msgtype": "markdown",
            "markdown": {
                "content": f"## {title}\n\n{message}"
            }
        }
        
        try:
            response = requests.post(
                self.webhook_url,
                json=payload,
                headers={'Content-Type': 'application/json'},
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                if result.get('errcode') == 0:
                    return True, "消息发送成功"
                else:
                    return False, f"企业微信API错误: {result.get('errmsg', '未知错误')}"
            else:
                return False, f"HTTP错误: {response.status_code}"
                
        except requests.exceptions.RequestException as e:
            return False, f"网络请求失败: {str(e)}"
        except json.JSONDecodeError as e:
            return False, f"JSON解析失败: {str(e)}"
    
    def send_log_summary(self, file_path, summary):
        """发送日志摘要"""
        time_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        
        message = f"""
**🚨 日志摘要通知**

**📁 文件:** `{file_path}`
**⏰ 时间:** `{time_str}`

**📋 摘要内容:**
```
{summary}
```
"""
        
        return self.send_message(message, "🚨 日志摘要通知")
    
    def send_error(self, file_path, error_msg):
        """发送错误通知"""
        time_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        
        message = f"""
**🚨 严重错误警报**

**📁 文件:** `{file_path}`
**⏰ 时间:** `{time_str}`

**💥 错误详情:**
```
{error_msg}
```
"""
        
        return self.send_message(message, "🚨 严重错误警报")

def main():
    """主函数"""
    if len(sys.argv) < 3:
        print("使用方法:")
        print("  python wechat_bot.py <webhook_url> <消息类型> [参数...]")
        print("")
        print("消息类型:")
        print("  summary <文件路径> <摘要内容>")
        print("  error <文件路径> <错误信息>")
        print("  message <标题> <消息内容>")
        sys.exit(1)
    
    webhook_url = sys.argv[1]
    msg_type = sys.argv[2]
    
    bot = WeChatBot(webhook_url)
    
    if msg_type == "summary":
        if len(sys.argv) != 5:
            print("错误: summary 类型需要文件路径和摘要内容")
            sys.exit(1)
        
        file_path = sys.argv[3]
        summary = sys.argv[4]
        success, result = bot.send_log_summary(file_path, summary)
        
    elif msg_type == "error":
        if len(sys.argv) != 5:
            print("错误: error 类型需要文件路径和错误信息")
            sys.exit(1)
        
        file_path = sys.argv[3]
        error_msg = sys.argv[4]
        success, result = bot.send_error(file_path, error_msg)
        
    elif msg_type == "message":
        if len(sys.argv) != 5:
            print("错误: message 类型需要标题和消息内容")
            sys.exit(1)
        
        title = sys.argv[3]
        message = sys.argv[4]
        success, result = bot.send_message(message, title)
        
    else:
        print(f"错误: 未知的消息类型 '{msg_type}'")
        sys.exit(1)
    
    if success:
        print(f"✅ {result}")
        sys.exit(0)
    else:
        print(f"❌ {result}")
        sys.exit(1)

if __name__ == "__main__":
    main()