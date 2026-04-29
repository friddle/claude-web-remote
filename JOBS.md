# JOBS - Claude Web Remote

## Priority Tasks

### 1. Webhook System Changes
- [ ] Only send webhook notifications when tasks **stop** (not during progress)
- [ ] Add Feishu webhook URL support
  - Support callback and notify functionality
  - Format messages properly for Feishu

### 2. Authentication Changes
- [ ] Change default auth order:
  1. First check `Authorization_$SESSION` header (session-specific)
  2. Fallback to `Authorization` header (global)
- [ ] Update client to send both headers:
  1. `Authorization_$SESSION` with session-specific token
  2. `Authorization` with global token
- [ ] Ensure username/password isolation:
  - Different sessions on same website should not impact each other
  - Each session has independent authentication

### 3. Debug Script
- [ ] Create `debug_local.sh` script
- [ ] Run clauded with specified host: `clauded.tools.yicoson.cn`
- [ ] Add `debug_local.sh` to `.gitignore`

## Implementation Notes

### Reference
- Claude Code hooks documentation: https://code.claude.com/docs/zh-TW/hooks
- Key events: `Stop`, `SubagentStop` for task completion

### Webhook Payload Format
Based on Claude Code hooks, the stop event includes:
```json
{
  "session_id": "string",
  "transcript_path": "string",
  "permission_mode": "string",
  "hook_event_name": "Stop",
  "stop_hook_active": boolean
}
```

### Feishu Webhook Format
```
msg_type: "text"
content: {
  "text": "message here"
}
```

### Authentication Headers
- Client sends: `Authorization_{SESSIONID}: token`
- Client sends: `Authorization: token`
- Server checks: `Authorization_{SESSIONID}` first, then `Authorization`
