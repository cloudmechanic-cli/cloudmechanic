<p align="center">
  <h1 align="center">CloudMechanic</h1>
  <p align="center"><strong>An OBD scanner for your AWS environment.</strong></p>
  <p align="center">Find cost leaks and security vulnerabilities in seconds — not hours.</p>
</p>

<p align="center">
  <a href="#installation"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/cloudmechanic-cli/cloudmechanic/releases"><img src="https://img.shields.io/github/v/release/cloudmechanic-cli/cloudmechanic?color=green" alt="Release"></a>
</p>

---

## The Problem

Every AWS account accumulates waste and risk over time:

- **Forgotten resources** — unattached EBS volumes, idle load balancers, unused Elastic IPs quietly burn money month after month.
- **Misconfigured security** — security groups left open to `0.0.0.0/0`, IAM users without MFA, public S3 buckets waiting to become the next headline.
- **Manual audits don't scale** — clicking through the AWS Console or writing one-off scripts is slow, error-prone, and never gets done consistently.

Teams discover these problems when the bill spikes or after a breach. By then, the damage is done.

## The Solution

**CloudMechanic** is a fast, single-binary CLI tool that scans your AWS account and delivers a color-coded, actionable report in your terminal — with exact Terraform code to fix every issue.

- **Fast** — runs all checks concurrently using goroutines. Full scans complete in under 1 second.
- **Zero config** — uses your existing AWS CLI credentials. No agents, no SaaS, no signup.
- **Actionable** — every finding includes the resource ID, a specific remediation step, and **ready-to-apply Terraform HCL**.
- **Extensible** — built on a `Scanner` interface. Adding a new check is one file and one line of registration.

```
$ cloudmechanic scan

=== CloudMechanic Scan Report ===

Security Issues (3):
  🔴 [CRITICAL] Security Group sg-0db0d4a51a974f36b (caf-bastion-sg) allows SSH (port 22) from 0.0.0.0/0
     Resource: sg-0db0d4a51a974f36b
     Fix:      Restrict SSH access to specific IP ranges or use AWS Systems Manager Session Manager instead.
  🔴 [CRITICAL] Security Group sg-09d33757e356808df (launch-wizard-2) allows SSH (port 22) from 0.0.0.0/0
     Resource: sg-09d33757e356808df
     Fix:      Restrict SSH access to specific IP ranges or use AWS Systems Manager Session Manager instead.

--------------------------------------------------
✅ Scan complete in 986ms
   Total issues: 3 (3 critical, 0 warnings)
```

## Quick Start

Get scanning in under 2 minutes:

```bash
# 1. Install
brew tap cloudmechanic-cli/tap && brew install cloudmechanic

# 2. Configure AWS credentials (skip if already configured)
aws configure

# 3. Scan
cloudmechanic scan

# 4. Or launch the interactive dashboard
cloudmechanic dashboard
```

**Linux (no Homebrew needed):**
```bash
curl -L https://github.com/cloudmechanic-cli/cloudmechanic/releases/download/v1.6.2/cloudmechanic_1.6.2_linux_amd64.tar.gz -o /tmp/cm.tar.gz \
  && sudo tar -xzf /tmp/cm.tar.gz -C /tmp/ \
  && sudo install -m 755 /tmp/cloudmechanic /usr/local/bin/cloudmechanic
```

## Prerequisites

### AWS Credentials

CloudMechanic needs valid AWS credentials to read your account resources. It **never modifies** anything — all operations are read-only.

**Option A — AWS CLI (recommended):**
```bash
aws configure
# Enter your Access Key ID, Secret Access Key, and default region
```

**Option B — Environment variables:**
```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"
```

**Option C — Named profiles** (for multi-account setups):
```bash
aws configure --profile production
cloudmechanic scan --profile production
```

### IAM Permissions

The simplest approach is to attach the **AWS-managed `ReadOnlyAccess` policy** to your IAM user or role. This grants all the permissions CloudMechanic needs.

If you prefer a minimal least-privilege policy:

<details>
<summary>Click to expand minimal IAM policy</summary>

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeVolumes",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeAddresses",
        "ec2:DescribeSnapshots",
        "ec2:DescribeRegions",
        "ec2:DescribeNatGateways",
        "ec2:DescribeRouteTables",
        "ec2:DescribeVpcs",
        "ec2:DescribeFlowLogs",
        "s3:ListAllMyBuckets",
        "s3:GetPublicAccessBlock",
        "s3:GetEncryptionConfiguration",
        "s3:GetBucketVersioning",
        "rds:DescribeDBInstances",
        "cloudwatch:GetMetricData",
        "dynamodb:ListTables",
        "dynamodb:DescribeContinuousBackups",
        "dynamodb:DescribeTable",
        "lambda:ListFunctions",
        "lambda:GetFunctionUrlConfig",
        "iam:ListUsers",
        "iam:ListMFADevices",
        "sts:GetCallerIdentity"
      ],
      "Resource": "*"
    }
  ]
}
```

</details>

### Build from Source (optional)

Only required if you're not using Homebrew or the pre-built binaries:

- **Go 1.24+**

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap cloudmechanic-cli/tap
brew install cloudmechanic
```

### Direct Binary (Linux / macOS)

```bash
# Linux amd64
curl -L https://github.com/cloudmechanic-cli/cloudmechanic/releases/download/v1.6.2/cloudmechanic_1.6.2_linux_amd64.tar.gz -o /tmp/cm.tar.gz \
  && sudo tar -xzf /tmp/cm.tar.gz -C /tmp/ \
  && sudo install -m 755 /tmp/cloudmechanic /usr/local/bin/cloudmechanic

# Linux arm64 (Graviton)
curl -L https://github.com/cloudmechanic-cli/cloudmechanic/releases/download/v1.6.2/cloudmechanic_1.6.2_linux_arm64.tar.gz -o /tmp/cm.tar.gz \
  && sudo tar -xzf /tmp/cm.tar.gz -C /tmp/ \
  && sudo install -m 755 /tmp/cloudmechanic /usr/local/bin/cloudmechanic
```

### Go Install

```bash
go install github.com/cloudmechanic-cli/cloudmechanic@latest
```

### Build from Source

```bash
git clone https://github.com/cloudmechanic-cli/cloudmechanic.git
cd cloudmechanic
go build -o cloudmechanic .
```

## Usage

### Run All Scanners

```bash
cloudmechanic scan
```

### Target a Specific Region

```bash
cloudmechanic scan --region us-west-2
```

### Scan All Regions

```bash
cloudmechanic scan --all-regions
```

### Use a Named AWS Profile

```bash
cloudmechanic scan --profile production
```

### JSON Output (for CI/CD pipelines)

```bash
cloudmechanic scan -o json
```

### CSV Output (for spreadsheets)

```bash
cloudmechanic scan -o csv > report.csv
```

### Self-Update

```bash
cloudmechanic upgrade
```

Checks the latest GitHub release, downloads the correct binary for your OS/arch, and replaces the current binary in-place. No Homebrew or package manager required.

### Check Version

```bash
cloudmechanic version
```

### Combine Flags

```bash
cloudmechanic scan --profile staging --region eu-west-1 -o json
```

## Interactive Dashboard (TUI)

Launch a full-screen terminal dashboard with real-time scanning, region filtering, severity filtering, live search, and instant Terraform remediation code:

```bash
cloudmechanic dashboard
cloudmechanic dashboard --all-regions
cloudmechanic dashboard --profile production
```

```
 ☁  CloudMechanic  ·  AWS Security Scanner
 ╭──────────────╮  ╭──────────────╮
 │  ⬡  Regions  │  │  ≡  Issues  │   ← Tab to switch. Active pane has bright border.
 ╰──────────────╯  ╰──────────────╯

 ╭────────────────────────╮  ╭──────────────────────────────────────────────────────╮
 │ REGIONS                │  │  SEVERITY    SERVICE    DESCRIPTION                  │
 │                        │  │                                                      │
 │  ◉ All Regions  3🔴 2🟡 │  │  🖥  EC2                                             │
 │    us-east-1    2🔴 1🟡 │  │   CRITICAL   EC2       sg-0db0d4a allows SSH 0.0.0.0│
 │    eu-west-1    1🔴 1🟡 │  │   WARNING    EC2       vol-0abc123 is unattached     │
 │                        │  │                                                      │
 │ SUMMARY                │  │  🪣  S3                                              │
 │  3  Critical            │  │   CRITICAL   S3        my-bucket has no PAB enabled │
 │  2  Warnings            │  │                                                     │
 │  5  Total               │  │                                                     │
 ╰────────────────────────╯  ╰──────────────────────────────────────────────────────╯
 ☁ us-east-1  1.2s   Tab Switch  ↑↓ Nav  ↵ Fix  F Filter  / Search  R Rescan  Q Quit
```

**Visual design:** Catppuccin Macchiato dark theme — the active pane has a bright blue border, the inactive pane is dimmed. The tab bar above the panes shows which side is focused.

### Dashboard Keybindings

| Key | Action |
|-----|--------|
| `Tab` | Switch focus between Regions sidebar and Issues list |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | **Open Terraform remediation view for the selected issue** |
| `F` | Cycle severity filter — All → Critical Only → Warnings Only |
| `/` | Live search across descriptions, resource IDs, and scanner names |
| `Esc` | Clear search |
| `R` | Re-run the scan |
| `Q` / `Ctrl+C` | Quit |

## Terraform Remediation View

Press `Enter` on any issue in the dashboard to open a full-screen code editor showing the exact Terraform HCL to fix it — with syntax highlighting.

```
  TERRAFORM REMEDIATION
  ─────────────────────────────────────────────────────────────────
  CRITICAL  VPC  VPCs Without Flow Logs
  Resource: vpc-0abc1234def567890         Region: us-east-1

  Enable VPC Flow Logs to CloudWatch
  Flow Logs capture metadata for every IP packet crossing your VPC.
  Essential for security forensics, intrusion detection, and compliance.

  ╭─ main.tf ──────────────────────────────────────────────────────╮
  │ resource "aws_iam_role" "flow_logs" {                          │
  │   name = "vpc-flow-logs-role"                                  │
  │                                                                │
  │   assume_role_policy = jsonencode({                            │
  │     Version = "2012-10-17"                                     │
  │     Statement = [{                                             │
  │       Effect    = "Allow"                                      │
  │       Action    = "sts:AssumeRole"                             │
  │       Principal = { Service = "vpc-flow-logs.amazonaws.com" }  │
  │     }]                                                         │
  │   })                                                           │
  │ }                                                              │
  │                                                                │
  │ resource "aws_flow_log" "fix" {                                │
  │   iam_role_arn    = aws_iam_role.flow_logs.arn                 │
  │   log_destination = aws_cloudwatch_log_group.flow_logs.arn     │
  │   traffic_type    = "ALL"                                      │
  │   vpc_id          = "<YOUR_VPC_ID>"                            │
  │ }                                                              │
  ╰────────────────────────────────────────────────────────────────╯

  [j/k] Scroll   [Esc / q] Back to Issues   [Ctrl+C] Quit
```

Terraform snippets are available for all 15 scanner types:

| Issue | Terraform Resource |
|---|---|
| Public S3 Buckets | `aws_s3_bucket_public_access_block` |
| S3 Without Encryption | `aws_s3_bucket_server_side_encryption_configuration` |
| S3 Without Versioning | `aws_s3_bucket_versioning` |
| VPCs Without Flow Logs | `aws_flow_log` + IAM role + CloudWatch log group |
| Open Security Groups (SSH) | `aws_security_group_rule` |
| Unused Elastic IPs | `aws_eip_association` |
| Unattached EBS Volumes | `aws_volume_attachment` |
| Old EBS Snapshots | `aws_dlm_lifecycle_policy` |
| Unused NAT Gateways | VPC module config |
| IAM Users Without MFA | `aws_iam_policy` with MFA deny condition |
| Idle RDS Instances | Stop/delete guidance |
| DynamoDB Without Backups | `point_in_time_recovery` block |
| DynamoDB Provisioned Capacity | `billing_mode = "PAY_PER_REQUEST"` |
| Lambda Deprecated Runtimes | `runtime` update |
| Lambda Public Function URLs | `authorization_type = "AWS_IAM"` |

## Current Scanners

| Scanner | Type | What It Finds |
|---------|------|---------------|
| **Unattached EBS** | Cost Leak | EBS volumes in `available` state (not attached to any instance) |
| **Open Security Groups** | Security | Security groups allowing `0.0.0.0/0` or `::/0` ingress on port 22 |
| **Public S3 Buckets** | Security | Buckets without full S3 Block Public Access enabled |
| **IAM Users Without MFA** | Security | IAM users with no MFA device enabled |
| **Idle RDS Instances** | Cost Leak | RDS instances with 0 connections over the last 7 days |
| **Unused Elastic IPs** | Cost Leak | Elastic IPs not associated with any resource ($3.60/mo each) |
| **Old EBS Snapshots** | Cost Leak | EBS snapshots older than 90 days |
| **DynamoDB Without Backups** | Security | Tables without Point-in-Time Recovery (PITR) |
| **DynamoDB Provisioned Capacity** | Cost Leak | Tables using provisioned mode that may benefit from on-demand |
| **Unused NAT Gateways** | Cost Leak | NAT Gateways not referenced in any route table (~$32/mo) |
| **VPCs Without Flow Logs** | Security | VPCs with no Flow Logs for network auditing |
| **Lambda Deprecated Runtimes** | Security | Functions running on EOL runtimes without security patches |
| **Lambda Public Function URLs** | Security | Functions with public URLs and no authentication |
| **S3 Buckets Without Encryption** | Security | Buckets with no default server-side encryption |
| **S3 Buckets Without Versioning** | Cost Leak | Buckets without versioning (risk of data loss) |

## Roadmap

- [x] Unused Elastic IPs
- [x] Public S3 Buckets
- [x] IAM Users without MFA
- [x] Idle RDS Instances (0 connections over 7 days)
- [x] Old EBS Snapshots (>90 days)
- [x] JSON / CSV output formats (`--output json`, `--output csv`)
- [x] Multi-region scanning (`--all-regions`)
- [x] DynamoDB backup & capacity checks
- [x] VPC Flow Logs & unused NAT Gateway checks
- [x] Lambda deprecated runtime & public URL checks
- [x] S3 encryption & versioning checks
- [x] Interactive TUI dashboard (`cloudmechanic dashboard`)
- [x] Two-pane explorer with region filtering, severity filter, and live search
- [x] Terraform remediation view — press Enter on any issue for ready-to-apply HCL
- [x] Self-update command (`cloudmechanic upgrade`)
- [x] Premium UI overhaul — Catppuccin Macchiato theme, visual tab bar, pill-shaped status bar, service emoji icons
- [ ] HTML report export
- [ ] Slack / webhook notifications
- [ ] Cost estimation per issue
- [ ] Custom severity thresholds

## Contributing

Contributions are welcome! CloudMechanic uses a simple `Scanner` interface — adding a new check is straightforward:

1. Create a new file in `internal/scanner/`
2. Implement the `Scanner` interface (`Name()` + `Scan()`)
3. Register it in `cmd/scan.go`
4. Submit a PR

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Built with Go, caffeine, and a healthy fear of surprise AWS bills.
</p>
