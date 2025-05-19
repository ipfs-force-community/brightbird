<h1 align="center">Bright Bird (重明鸟)</h1>
<p align="center">
  <a href="README-zh.md"><img src="https://img.shields.io/badge/简体中文-orange"></a>
  <a href="README.md"><img src="https://img.shields.io/badge/English-yellow"></a>
  <br>
</p>

## Introduction

Brightbird is an automated testing platform designed for Kubernetes environments. It simplifies the process of setting up, executing, and managing integration, regression, and system tests, especially for complex systems like Venus. Through a graphical interface, users can define test flows, manage dependencies, and automate the deployment of necessary resources within a Kubernetes cluster.

## Features

Key features include:

- Visual test flow orchestration (drag-and-drop UI)
- Deep Kubernetes integration (automated deployment and management of K8s resources)
- Plugin architecture (supports Deployer and Exec plugin extensions)
- Parameter and dependency management (global parameters, inter-node data transfer)
- Flexible task triggering mechanisms (supports Cron scheduling and Webhooks)
- Status monitoring and log management (real-time status tracking, centralized logging based on Fluent Bit and MongoDB)
- Web-based user interface

## Architecture

System architecture overview:

- Backend: Developed in Go, using Gin/Fx framework, MongoDB database
- Frontend: Based on TypeScript, using Vue/React framework and graph libraries
- Test Runner: Developed in Go
- Plugin System: Supports extensible functionality
- Logging System: Based on Fluent Bit, with customized MongoDB output
- Database: MongoDB

## Core Concepts

Core concept definitions:

- **Test Flow**: Overall definition of the test process
- **Node**: Execution unit in the test flow, associated with plugins
- **Plugin**: Includes Deployer and Exec modules, providing specific functionality
- **Job**: Execution instance of a test flow, can be triggered via Cron or Webhook
- **Parameters**: Used for configuration and data transfer between nodes
- **Trigger**: Mechanism to start jobs (such as Cron, Webhook)

## Quick Start

[Brightbird User Guide](docs/user-guide.md)：

1. Upload necessary plugins
2. Use the graphical editor to create test flows
3. Configure nodes and link dependencies/parameters
4. Create tasks to run test flows
5. Trigger tasks (manually or via schedule/webhook)
6. Monitor execution and view logs

## Plugin Development

Brightbird supports two types of plugins: **Deployer** (for setting up environments) and **Exec** (for running test logic). Plugins are written in Go and define their inputs/outputs using structs.

- Detailed plugin development guidelines please refer to [dev-plugin](docs/dev-plugin.md)
- **Build Command** (for Linux compatibility):

  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o <plugin_name> main.go
  ```

## [Deploy](docs/deploy.md)

## Technology Stack

List the main technologies used:

- **Backend**: Go, Fx, MongoDB Go Driver, Kubernetes Client Go
- **Frontend**: TypeScript, Vue.js (likely), Axios, Graph Library (X6/G6 likely)
- **Database**: MongoDB
- **Logging**: Fluent Bit
- **Environment**: Kubernetes, Docker
- **Others**: JSON Schema, YAML
