# Minerva Olive

Minerva Olive is a configuration and feature flag server built for Minerva platform,
but can be used by any other platform.

[![macOS Test](https://github.com/sy-software/minerva-olive/workflows/Test/badge.svg)](https://github.com/sy-software/minerva-olive/actions)
[![codecov](https://codecov.io/gh/sy-software/minerva-olive/branch/main/graph/badge.svg?token=ATZPRNEL7Y)](https://codecov.io/gh/sy-software/minerva-olive)
![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)

## Config File

To run this project you'll need to provide a config file in json format.

```json
{
    // Redis connection configurations
    "redisConfig": {
        // Database host, default: 127.0.0.1
        "host": "127.0.0.1",
        // Database port, default: 6379
        "port": "6379",
        // Database selected, default: 1
        "db": 1,
        // Database username. Omit if the DB have no authentication
        "username": "",
        // Database password. Omit if the DB have no authentication
        "password": "",
        // Maximum number of retries before giving up, default: 3
        "maxRetries": 3,
        // Database connection timeout, default: 10 seconds
        "connectionTimeout": 10,
        // Database read timeout, default: 1s
        "readTimeout": 1,
        // Max number of socket connections, default: 10
        "poolSize": 10
    },
    // TTL for stored cache, default (-1) infinite
    "cacheTTL": -1,
    // Server bind IP default 0.0.0.0
    "host": "0.0.0.0",
    // Server bind port default 8080
    "port": 8080
}

```
## Environment Variables

You can use the following environment varibles to configure the server:

```sh
# Set this variable to get human readable log output
CONSOLE_OUTPUT=1
# Set this to the desired log level. Default: INFO
LOG_LEVEL=DEBUG
# Set this to release to hide gin debug logs
GIN_MODE=release
# Set this to use with AWS Secret manager
AWS_REGION=us-east-1
# Path to configuration file. Default: ./config.json
CONFIG_FILE=./config.json
```


## Run Locally

Clone the project

```bash
  git clone https://github.com/sy-software/minerva-olive
```

Go to the project directory

```bash
  cd minerva-olive
```

Install dependencies

```bash
  make deps
```

Start the server

```bash
  make run
```


## Deployment

Build and deploy docker image

```bash
  make docker-build
```
