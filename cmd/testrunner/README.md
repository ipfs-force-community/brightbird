# cmd/testrunner

Testrunner is responsible for executing tests, managing test environments, and reporting results.

## Overview

Testrunner consists of several key components:

- **`api_controller.go`**: Defines the API controller that exposes endpoints for managing and interacting with testrunner. It uses the Gin framework for routing and handling HTTP requests.
- **`clean.go`**: Implements a cleanup mechanism to ensure proper resource release after test runs, even in case of errors. It uses a series of cleanup functions to handle various cleanup tasks.
- **`node_exec.go`**: Contains the logic for executing individual nodes in the test flow. It handles parsing input values, executing commands, and managing each node's state.
- **`value_resolve.go`**: Provides functions for resolving values in JSON structures, enabling dynamic configuration and data injection during test execution.

## Features

Testrunner provides the following core functionality:

- **API Management**: Exposes APIs for starting, stopping, and monitoring test runs.
- **Environment Setup**: Creates and manages test environments including Kubernetes resources and other dependencies.
- **Test Execution**: Executes test flows defined as interconnected node graphs.
- **Value Resolution**: Parses dynamic values in test configurations.
- **Cleanup**: Ensures proper resource release after test execution.
- **Logging**: Provides detailed logging of test execution for debugging and analysis.

## Code Structure

- **`APIController`**: Manages API endpoints and interacts with testrunner's core logic.
- **`Cleaner`**: Handles resource cleanup after test execution.
- **`resolveInputValue`**: Parses input values for test nodes based on schema and environment context.
- **`IterJSON`**: Dynamically iterates and resolves values in JSON structures.

## Parameter Usage

Testrunner can be configured with the following command-line parameters:

- `--config`: Path to TOML configuration file.
- `--plugins`: Path to plugin storage directory.
- `--namespace`: Kubernetes namespace to use (default: `default`).
- `--mysql`: MySQL connection string template (e.g. `xxx%sxxxx`).
- `--mongoUrl`: MongoDB connection URL (default: `mongodb://localhost:27017`).
- `--dbName`: MongoDB database name (default: `testplateform`).
- `--timeout`: Timeout for test units in minutes (default: `0`, no timeout).
- `--taskId`: ID of task to run.
- `--retry`: Number of retries for failed tasks.
- `--registry`: Use private registry for Docker images.
- `--listen`: Address to listen for API requests (default: `0.0.0.0:5682`).
- `--globalParams`: Global parameters in JSON format (default: `{}`).
