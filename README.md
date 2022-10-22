# Argusyes-backend

A performance monitor use ssh protocol

## Websocket interface example

### Connect establish

```javascript
url = 'ws://localhost:9097/monitor';
c = new WebSocket(url); 
```

### Monitor

request example

```json
{
  "id": "2138a74f91264b1",
  "method": "ssh.start_monitor",
  "params": [
    {
      "port": 22,
      "host": "10.128.248.93",
      "user": "cc",
      "passwd": "chenchen"
    }
    // more ssh connect
  ]
}
```

response example

```json
{
  "id": "same as request",
  "result": null,
  "error": null
  // if no err else a string descibe the error
}
```

### UnMonitor

request example

```json
{
  "id": "32190b82109a23",
  "method": "ssh.stop_monitor",
  "params": [
    {
      "port": 22,
      "host": "10.128.248.93",
      "user": "cc"
    }
    // more ssh connect
  ]
}
```

response example

```json
{
  "id": "same as request",
  "result": null,
  "error": null
  // if no err else a string descibe the error
}
```

### Notification

#### CPUInfoNotification

```json
{
  "id": null,
  "method": "ssh.notification",
  "params": [
    {
      "ssh_key": "cc@10.128.248.93:22",
      "event": "cpu_info",
      "message": {
        "ssh_key": "cc@10.128.248.93:22",
        "processor_num": 8,
        "cpu_info": [
          {
            "processor": 0,
            "model_name": " Intel(R) Xeon(R) CPU E3-1265L v3 @ 2.50GHz"
          },
          {
            "processor": 1,
            "model_name": " Intel(R) Xeon(R) CPU E3-1265L v3 @ 2.50GHz"
          },
          {
            "processor": 2,
            "model_name": " Intel(R) Xeon(R) CPU E3-1265L v3 @ 2.50GHz"
          }
          // more cpu info
        ]
      }
    }
  ]
}
```