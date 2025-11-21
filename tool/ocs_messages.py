#!/usr/bin/env python3

import json, os, sys
import argparse

STORAGE_ROOT = os.path.expanduser("~/.local/share/opencode/storage")

def load_json(path):
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)

def find_files_with_text(text):
    files = []
    for root, dirs, filenames in os.walk(STORAGE_ROOT):
        for filename in filenames:
            if filename.endswith('.json'):
                path = os.path.join(root, filename)
                try:
                    with open(path, 'r', encoding='utf-8') as f:
                        content = f.read()
                        if text in content:
                            files.append(path)
                except:
                    pass
    return files

def find_parts(message_id):
    parts_dir = os.path.join(STORAGE_ROOT, "part", message_id)
    parts = []
    if os.path.exists(parts_dir):
        for filename in os.listdir(parts_dir):
            if filename.endswith('.json'):
                path = os.path.join(parts_dir, filename)
                try:
                    part_data = load_json(path)
                    if isinstance(part_data, dict):
                        parts.append(part_data)
                except:
                    pass
    parts.sort(key=lambda p: p.get("order", 0))
    return parts

def recompose_session(session_id):
    try:
        messages_dir = os.path.join(STORAGE_ROOT, "message", session_id)
        messages = []
        if os.path.exists(messages_dir):
            for filename in os.listdir(messages_dir):
                if filename.endswith('.json'):
                    path = os.path.join(messages_dir, filename)
                    try:
                        msg_data = load_json(path)
                        if isinstance(msg_data, dict) and msg_data.get("id", "").startswith("msg_"):
                            message = {
                                "id": msg_data.get("id"),
                                "role": msg_data.get("role", "user"),
                                "createdAt": msg_data.get("time", {}).get("created"),
                                "parts": [],
                                "metadata": {}
                            }
                            parts_data = find_parts(message["id"])
                            if not parts_data:
                                # Fallback to summary
                                text = ""
                                if message["role"] == "user":
                                    summary = msg_data.get("summary", {})
                                    title = summary.get("title", "")
                                    body = summary.get("body", "")
                                    text = title
                                    if body:
                                        text += "\n" + body
                                elif message["role"] == "assistant":
                                    if "error" in msg_data:
                                        text = msg_data["error"].get("message", "")
                                if text:
                                    message["parts"] = [{"type": "text", "content": text, "order": 0}]
                            else:
                                message["parts"] = parts_data
                            messages.append(message)
                    except:
                        pass
        messages.sort(key=lambda m: m.get("createdAt") or 0)
        return {"id": session_id, "messages": messages}
    except KeyboardInterrupt:
        print("Interrupted")
        sys.exit(1)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('session_id')
    parser.add_argument('-d', '--storage-root', default=os.path.expanduser("~/.local/share/opencode/storage"))
    args = parser.parse_args()
    STORAGE_ROOT = args.storage_root
    session_id = args.session_id
    try:
        result = recompose_session(session_id)
        for msg in result["messages"]:
            for part in msg["parts"]:
                content = part.get("content", "") or part.get("text", "")
                if content:
                    part_type = part.get("type", "text")
                    if part_type == "reasoning":
                        label = "Reasoning"
                    elif part_type == "tool":
                        label = "Tool"
                    elif part_type == "text" and content.startswith("<file>"):
                        label = "File"
                        print(f"> # {label}:\n\n```\n{content}\n```")
                        print()
                        continue
                    elif part_type == "text" and "Called the" in content and "tool" in content:
                        label = "Tool"
                    else:
                        label = msg["role"].capitalize()
                    print(f"> # {label}:\n\n {content}")
                    print()
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
