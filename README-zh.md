<h1 align="center">Bright Bird (重明鸟)</h1>
<p align="center">
  <a href="README-zh.md"><img src="https://img.shields.io/badge/简体中文-orange"></a>
  <a href="README.md"><img src="https://img.shields.io/badge/English-yellow"></a>
  <br>
</p>

## 简介

Brightbird 是一个为 Kubernetes 环境设计的自动化测试平台。它简化了集成、回归和系统测试的设置、执行和管理过程，特别适用于像 Venus 这样的复杂系统。通过图形界面，用户可以定义测试流程、管理依赖关系，并在 Kubernetes 集群中自动部署所需资源。

## 功能特性

主要功能包括：

* 可视化测试流程编排（拖拽式UI）
* 深度集成Kubernetes（自动化部署和管理K8s资源）
* 插件化架构（支持Deployer和Exec插件扩展）
* 参数化和依赖管理（全局参数、节点间数据传递）
* 灵活的任务触发机制（支持Cron调度和Webhooks）
* 状态监控与日志管理（实时状态跟踪，基于Fluent Bit和MongoDB的集中式日志管理）
* 基于Web的用户界面

## 系统架构

系统架构概览：

* 后端：使用Go语言开发，框架为Gin/Fx，数据库为MongoDB
* 前端：基于TypeScript，使用Vue/React框架和图形库
* 测试运行器：使用Go语言开发
* 插件系统：支持扩展功能
* 日志系统：基于Fluent Bit，定制化MongoDB输出
* 数据库：MongoDB

## 核心概念

核心概念定义：

* **测试流程（Test Flow）**：测试过程的整体定义
* **节点（Node）**：测试流程中的执行单元，与插件相关联
* **插件（Plugin）**：包括Deployer和Exec模块，提供特定功能
* **任务（Job）**：测试流程的执行实例，可通过Cron或Webhook触发
* **参数（Parameters）**：用于配置和在节点间传递数据的输入/输出
* **触发器（Trigger）**：启动任务的机制（如Cron、Webhook）

## 快速开始

[Brightbird使用教程](docs/user-guide.md)：

1. 访问Web界面
2. 上传必要的插件
3. 使用图形编辑器创建测试流程
4. 配置节点并链接依赖关系/参数
5. 创建任务以运行测试流程
6. 触发任务（手动或通过计划/webhook）
7. 监控执行过程并查看日志

## 插件开发

Brightbird支持两种类型的插件：**Deployer**（用于环境设置）和**Exec**（用于运行测试逻辑）。插件使用Go语言编写，并使用结构体定义输入/输出。

* 详细的插件开发指南请参考 [docs/dev-plugin.md](docs/dev-plugin.md)
* **构建命令**（用于Linux兼容性）：

  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o <plugin_name> main.go
  ```

## [系统部署](docs/deploy.md)

## 技术栈

主要使用的技术：

* **后端**：Go、Fx、MongoDB Go Driver、Kubernetes Client Go
* **前端**：TypeScript、Vue.js（可能）、Axios、图形库（可能是X6/G6）
* **数据库**：MongoDB
* **日志系统**：Fluent Bit
* **环境**：Kubernetes、Docker
* **其他**：JSON Schema、YAML
