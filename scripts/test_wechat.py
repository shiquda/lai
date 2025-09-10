#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
企业微信通知测试脚本
"""

import requests
import json
import sys

def test_wechat_webhook(webhook_url):
    """测试企业微信 Webhook"""
    if not webhook_url:
        print("❌ 请提供 Webhook URL")
        return False
    
    # 测试消息
    payload = {
        "msgtype": "markdown",
        "markdown": {
            "content": """
## 🧪 企业微信测试消息

✅ **Lai 企业微信通知测试成功！**

**测试内容：**
- 🤖 机器人状态：正常
- 📡 Webhook 连接：正常
- 💬 消息格式：Markdown

**时间：** 2025-09-10

---
*此消息由 Lai 日志监控系统发送*
            """
        }
    }
    
    try:
        response = requests.post(
            webhook_url,
            json=payload,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )
        
        print(f"📡 HTTP 状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print(f"📋 响应内容: {json.dumps(result, ensure_ascii=False, indent=2)}")
            
            if result.get('errcode') == 0:
                print("✅ 企业微信 Webhook 测试成功！")
                return True
            else:
                print(f"❌ 企业微信 API 错误: {result.get('errmsg', '未知错误')}")
                return False
        else:
            print(f"❌ HTTP 错误: {response.status_code}")
            print(f"📋 响应内容: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ 网络请求失败: {str(e)}")
        return False
    except json.JSONDecodeError as e:
        print(f"❌ JSON 解析失败: {str(e)}")
        return False

def main():
    """主函数"""
    if len(sys.argv) != 2:
        print("使用方法:")
        print("  python test_wechat.py <webhook_url>")
        print("")
        print("示例:")
        print("  python test_wechat.py https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY")
        sys.exit(1)
    
    webhook_url = sys.argv[1]
    
    print("🧪 开始测试企业微信 Webhook...")
    print(f"📡 Webhook URL: {webhook_url[:50]}...")
    print("")
    
    success = test_wechat_webhook(webhook_url)
    
    if success:
        print("")
        print("🎉 测试完成！企业微信通知配置正确。")
        print("💡 现在你可以在 Lai 中使用企业微信通知了。")
    else:
        print("")
        print("⚠️  测试失败！请检查：")
        print("1. Webhook URL 是否正确")
        print("2. 网络连接是否正常")
        print("3. 企业微信群机器人是否正常工作")
    
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()