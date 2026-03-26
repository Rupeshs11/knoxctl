<div align="center">
  <h1>☸️ Knoxctl</h1>
  <p><b>A lightning-fast, lightweight CLI for Kubernetes operations.</b></p>

[![Go Report Card](https://goreportcard.com/badge/github.com/Rupeshs11/knoxctl)](https://goreportcard.com/report/github.com/Rupeshs11/knoxctl)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-lightgrey)](#)

</div>

<br>

**knoxctl** is a high-performance Kubernetes CLI tool built in **Go** using the Cobra framework and `client-go`.

Designed as a streamlined, developer-friendly alternative to `kubectl`, it strips away the noise and focuses on the core operations you actually use every day: **Get**, **Apply**, and **Delete**. It seamlessly parses your local `~/.kube/config`, directly handles raw YAML manifests via dynamic client rendering, and presents the data in a beautiful, formatted tabular view.

---

## ✨ Why Knoxctl? (Features)

🚀 **Blazing Fast**: Compiled directly to single-binary machine code using Go.  
💻 **Cross-Platform**: Ready-to-use binaries for Windows and Linux (WSL).  
🛠️ **Server-Side Apply (SSA)**: Uses native dynamic clients to robustly create, update, and patch complex YAML structures without needing hardcoded schemas.  
🎯 **"All" Commands**: Bulk-target resources to save time! Supports `get all`, `apply -f all`, and `delete all` within any namespace.  
🧹 **Clean Visuals**: Tabular output strictly focused on the critical metadata (Status, Age, Replicas, IPs).

---

## 📦 Installation

There are two ways to get **knoxctl** running on your system.

> **Note:** Make sure you have `kubectl` or your cluster properly configured with a valid `~/.kube/config`. **knoxctl** automatically reads your environment!

### Option 1: Download Pre-built Binary (Recommended)

**🐧 Linux / Ubuntu (WSL)**

```bash
# Download the latest Linux binary
wget https://github.com/Rupeshs11/knoxctl/releases/download/v1.0.0/knoxctl-linux-amd64 -O knoxctl

# Make it executable
chmod +x knoxctl

# Move it to your local bin for global access
sudo mv knoxctl /usr/local/bin/knoxctl

# Verify
knoxctl --help
```

**🪟 Windows (PowerShell)**

```powershell
# 1. Download knoxctl.exe from the GitHub Releases page
# 2. Move it to a directory that is in your PATH. For example:
Move-Item -Path ".\Downloads\knoxctl.exe" -Destination "C:\Windows\System32\knoxctl.exe"

# Verify
knoxctl --help
```

<details>
<summary><b>Option 2: Build from Source (For Developers)</b></summary>
<br>

If you have Go installed and want to compile it yourself:

```bash
# 1. Clone the repository
git clone https://github.com/Rupeshs11/knoxctl.git
cd knoxctl

# 2. Download Go modules
go mod tidy

# 3. Build the binary
go build -o knoxctl .       # For Linux/macOS
go build -o knoxctl.exe .   # For Windows

# 4. Move to your PATH
```

</details>

---

## 🚀 Commands & Usage

```text
knoxctl [command] [flags]
```

### 🔍 1. Read Operations (`get`)

Instantly fetch the status of your workloads.

| Command                   | Description                                  |
| ------------------------- | -------------------------------------------- |
| `knoxctl get pods`        | List pods in a namespace                     |
| `knoxctl get deployments` | List deployments in a namespace              |
| `knoxctl get svc`         | List services in a namespace                 |
| `knoxctl get nodes`       | List all active nodes in the cluster         |
| `knoxctl get ns`          | List all namespaces                          |
| **`knoxctl get all`**     | **List all pods, deployments, and services** |

**📸 In Action:**

<div align="center">
  <p><i>Getting detailed pod, deployment, and service metadata across all namespaces:</i></p>
  <img src="assets/get%20all%20from%20all%20namespaces.png" alt="Get All Namespaces" width="800"/>
  <br><br>
  <p><i>Fetching specific endpoints and services:</i></p>
  <img src="assets/get%20services.png" alt="Services" width="800"/>
</div>

---

### 🏗️ 2. Write Operations (`apply`)

Deploy resources directly from raw YAML files, effortlessly.

| Command                    | Description                                   |
| -------------------------- | --------------------------------------------- |
| `knoxctl apply -f <file>`  | Apply or create resources from a single YAML  |
| **`knoxctl apply -f all`** | **Apply all YAML files in current directory** |

**📸 In Action:**

<div align="center">
  <p><i>Deploying an entire directory of YAML configurations at once:</i></p>
  <img src="assets/apply%20all.png" alt="Apply All" width="800"/>
  <br><br>
  <p><i>Creating a specific deployment:</i></p>
  <img src="assets/Create%20and%20get%20deployment.png" alt="Apply Deployment" width="800"/>
</div>

---

### 💥 3. Destroy Operations (`delete`)

Quickly terminate workloads or teardown entire namespaces.

| Command                       | Description                               |
| ----------------------------- | ----------------------------------------- |
| `knoxctl delete <res> <name>` | Delete a specific resource                |
| `knoxctl delete -f <file>`    | Delete resources defined in a YAML file   |
| **`knoxctl delete all`**      | **Delete all pods, deployments, and svc** |

**📸 In Action:**

<div align="center">
  <p><i>Force cleaning an entire namespace (Deployments, Pods, Services):</i></p>
  <img src="assets/delete%20all%20in%20namespace.png" alt="Delete All Namespace" width="800"/>
  <br><br>
  <p><i>Tearing down a specific deployment:</i></p>
  <img src="assets/delete%20deployment.png" alt="Delete Deployment" width="800"/>
</div>

---

## 🏁 Global Flags

Append these to **any** command to control your execution context:

| Flag                     | Description                                                       |
| ------------------------ | ----------------------------------------------------------------- |
| `-n, --namespace <name>` | Target a specific namespace (default: "default")                  |
| `-A, --all-namespaces`   | List or target resources across **all** namespaces                |
| `--kubeconfig <path>`    | Provide a custom kubeconfig file path (default: `~/.kube/config`) |

---

