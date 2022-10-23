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
      "event": "cpu_info",
      "message": {
        "ssh_key": "cc@10.128.248.93:22",
        "cpu_info": {
          "0": {
            "cpu_core_info": {
              "0": {
                "cpu_processor_info": {
                  "0": {
                    "processor": 0,
                    "CPUMHz": 800,
                    "apicid": 0
                  },
                  "4": {
                    "processor": 4,
                    "CPUMHz": 800,
                    "apicid": 1
                  }
                },
                "core_id": 0
              },
              "1": {
                "cpu_processor_info": {
                  "1": {
                    "processor": 1,
                    "CPUMHz": 800,
                    "apicid": 2
                  },
                  "5": {
                    "processor": 5,
                    "CPUMHz": 800,
                    "apicid": 3
                  }
                },
                "core_id": 1
              },
              "2": {
                "cpu_processor_info": {
                  "2": {
                    "processor": 2,
                    "CPUMHz": 983.36,
                    "apicid": 4
                  },
                  "6": {
                    "processor": 6,
                    "CPUMHz": 800,
                    "apicid": 5
                  }
                },
                "core_id": 2
              },
              "3": {
                "cpu_processor_info": {
                  "3": {
                    "processor": 3,
                    "CPUMHz": 2066.414,
                    "apicid": 6
                  },
                  "7": {
                    "processor": 7,
                    "CPUMHz": 800,
                    "apicid": 7
                  }
                },
                "core_id": 3
              }
            },
            "vendor_id": "GenuineIntel",
            "cpu_family": "6",
            "model": "60",
            "model_name": "Intel(R) Xeon(R) CPU E3-1265L v3 @ 2.50GHz",
            "stepping": "3",
            "cache_size": "8192 KB",
            "physical_id": 0,
            "siblings": 8,
            "cpu_cores": 4,
            "fpu": true,
            "fpu_exception": false,
            "bogomips": 4988.55,
            "cl_flush_size": 64,
            "cache_alignment": 64,
            "address_sizes": "39 bits physical, 48 bits virtual"
          }
        }
      }
    }
  ]
}
```