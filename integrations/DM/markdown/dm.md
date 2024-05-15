# DM

dm 监控采集插件，核心原理就是连到 dm 实例，执行一些 sql，解析输出内容，整理为监控数据上报。

## Configuration

```toml
# collect interval
interval = 15

[[instances]]
address = "ip:5236"
username = "your username"
password = "your password"
dbname = ""

# set tls=custom to enable tls
parameters = "tls=false"

# timeout
timeout_seconds = 3

# interval = global.interval * interval_times
interval_times = 1

# important! use global unique string to specify instance
labels = { instance="达梦数据库", address="ip:5236" }

# Optional TLS Config
use_tls = false
tls_min_version = "1.2"
tls_ca = "/etc/jditms_agent/ca.pem"
tls_cert = "/etc/jditms_agent/cert.pem"
tls_key = "/etc/jditms_agent/key.pem"
# Use TLS but skip chain & host verification
insecure_skip_verify = true

```

## 监控多个实例
大家最常问的问题是如何监控多个mysql实例，实际大家对toml配置学习一下就了解了，`[[instances]]` 部分表示数组，是可以出现多个的，address参数支持通过unix路径连接 所以，举例：

```toml
[[instances]]
address = "ip:5236"
username = "your username"
password = "your password"
dbname = ""
labels = { instance="达梦数据库", address="ip:5236" }

[[instances]]
address = "ip:5236"
username = "your username"
password = "your password"
dbname = ""
labels = { instance="达梦数据库", address="ip:5236" }

[[instances]]
address = "ip:5236"
username = "your username"
password = "your password"
dbname = ""
labels = { instance="达梦数据库", address="ip:5236" }
```
