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
  "method": "ssh.startMonitor",
  "params": [
    {
      "port": 22,
      "host": "10.128.248.93",
      "user": "cc",
      "passwd": "chenchen"
    },
    {
      "port": 10022,
      "host": "10.112.230.222",
      "user": "chenchen",
      "passwd": "chenchen"
    }
  ]
}
```

response example

```json
{
  "id": "same as request",
  "result": [
    {
      "port": 22,
      "host": "10.128.248.93",
      "user": "cc",
      "monitor": true,
      "error": null
    },
    {
      "port": 10022,
      "host": "10.112.230.222",
      "user": "chenchen",
      "monitor": true,
      "error": null
    }
  ],
  "error": null
}
```

### UnMonitor

request example

```json
{
  "id": "32190b82109a23",
  "method": "ssh.stopMonitor",
  "params": [
    {
      "port": 22,
      "host": "10.128.248.93",
      "user": "cc"
    },
    {
      "port": 10022,
      "host": "10.112.230.222",
      "user": "chenchen",
      "passwd": "chenchen"
    }
  ]
}
```

response example

```json
{
  "id": "lll",
  "error": null,
  "result": [
    {
      "port": 22,
      "host": "10.128.248.93",
      "user": "cc",
      "unMonitor": true,
      "error": null
    },
    {
      "port": 10022,
      "host": "10.112.230.222",
      "user": "chenchen",
      "unMonitor": true,
      "error": null
    }
  ]
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
      "event": "cpuInfo",
      "message": {
        "port": 22,
        "host": "10.128.248.93",
        "user": "cc",
        "cpuInfo": {
          "0": {
            "cpuCoreInfo": {
              "0": {
                "cpuProcessorInfo": {
                  "0": {
                    "processor": 0,
                    "CPUMHz": 3353.203,
                    "apicid": 0
                  },
                  "4": {
                    "processor": 4,
                    "CPUMHz": 3349.356,
                    "apicid": 1
                  }
                },
                "coreId": 0
              },
              "1": {
                "cpuProcessorInfo": {
                  "1": {
                    "processor": 1,
                    "CPUMHz": 3352.969,
                    "apicid": 2
                  },
                  "5": {
                    "processor": 5,
                    "CPUMHz": 3350.782,
                    "apicid": 3
                  }
                },
                "coreId": 1
              },
              "2": {
                "cpuProcessorInfo": {
                  "2": {
                    "processor": 2,
                    "CPUMHz": 3092.994,
                    "apicid": 4
                  },
                  "6": {
                    "processor": 6,
                    "CPUMHz": 1400,
                    "apicid": 5
                  }
                },
                "coreId": 2
              },
              "3": {
                "cpuProcessorInfo": {
                  "3": {
                    "processor": 3,
                    "CPUMHz": 3352.634,
                    "apicid": 6
                  },
                  "7": {
                    "processor": 7,
                    "CPUMHz": 3349.121,
                    "apicid": 7
                  }
                },
                "coreId": 3
              }
            },
            "vendorId": "GenuineIntel",
            "cpuFamily": "6",
            "model": "60",
            "modelName": "Intel(R) Xeon(R) CPU E3-1265L v3 @ 2.50GHz",
            "stepping": "3",
            "cacheSize": "8192 KB",
            "physicalId": 0,
            "siblings": 8,
            "cpuCores": 4,
            "fpu": true,
            "fpuException": false,
            "bogomips": 4988.55,
            "clFlushSize": 64,
            "cacheAlignment": 64,
            "addressSizes": "39 bits physical, 48 bits virtual"
          }
        }
      }
    }
  ]
}
```