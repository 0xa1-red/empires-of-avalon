## Release v0.16.4 (2023-07-17)

### Features

#### EOA-39
* 103fe23 Upgrade ProtoActor

#### EOA-37
* b8a75a7 Add version information, automatically use version in instrumentation

#### EOA-15
* d5e1b40 Huge instrumentation rework
* 8058a9c Add metrics pipeline, serve promhttp

#### Other
* ba11ff8 Add missing traces, change logging level
* 65d0021 Add jaeger to docker compose

### Fixes

#### EOA-40
* 879f042 Add CMD to healthcheck test command

#### EOA-36
* 48eb410 Add ignore file for chart cache

#### Other
* 6b5ed61 Fix default empty reqID, log message
* 608e040 Fix server instantiation