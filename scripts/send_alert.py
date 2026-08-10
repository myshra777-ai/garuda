import os
import requests

TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
TELEGRAM_CHAT_ID = os.getenv("TELEGRAM_CHAT_ID")

def send_alert(message, commit_hash=None, job_url=None):
    if not TELEGRAM_BOT_TOKEN or not TELEGRAM_CHAT_ID:
        print("Telegram credentials unconfigured.")
        return

    alert = f"🚨 *GARUDA SYSTEM ALERT*\n\n{message}"
    if commit_hash:
        alert += f"\n📌 Commit: `{commit_hash}`"
    if job_url:
        alert += f"\n🔗 [View Job]({job_url})"

    url = f"https://api.telegram.org/bot{TELEGRAM_BOT_TOKEN}/sendMessage"
    payload = {
        "chat_id": TELEGRAM_CHAT_ID,
        "text": alert,
        "parse_mode": "Markdown",
        "disable_web_page_preview": True
    }

    try:
        response = requests.post(url, json=payload, timeout=5)
        if response.status_code == 200:
            print("✅ Alert sent!")
        else:
            print(f"Telegram error: {response.text}")
    except Exception as e:
        print(f"Alert failed: {e}")