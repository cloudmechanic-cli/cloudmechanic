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

**CloudMechanic** is a fast, single-binary CLI tool that scans your AWS account and delivers a color-coded, actionable report in your terminal.

- **Fast** — runs all checks concurrently using goroutines. Full scans complete in under 1 second.
- **Zero config** — uses your existing AWS CLI credentials. No agents, no SaaS, no signup.
- **Actionable** — every finding includes the resource ID and a specific remediation step.
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
  🔴 [CRITICAL] Security Group sg-02e017ca8231040c6 (launch-wizard-1) allows SSH (port 22) from 0.0.0.0/0
     Resource: sg-02e017ca8231040c6
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

### Go Install

```bash
go install github.com/cloudmechanic-cli/cloudmechanic@latest
```

### Download Binary

Grab the latest release for your platform from the [Releases](https://github.com/cloudmechanic-cli/cloudmechanic/releases) page.

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

### Check Version

```bash
cloudmechanic version
```

### Combine Flags

```bash
cloudmechanic scan --profile staging --region eu-west-1 -o json
```

## Interactive Dashboard (TUI)

Launch a full-screen terminal dashboard with real-time scanning, region sidebar, and expandable issue details:

```bash
cloudmechanic dashboard
```

```bash
cloudmechanic dashboard --all-regions
cloudmechanic dashboard --profile production
```

```
  ______  __                    __  __  ___              __                   _
 / ____/ / /____   __  __ ____/ / /  |/  /___   _____  / /_   ____ _ ____   (_)_____
/ /     / // __ \ / / / // __  / / /|_/ // _ \ / ___/ / __ \ / __ '// __ \ / // ___/
/ /___ / // /_/ // /_/ // /_/ / / /  / //  __// /__  / / / // /_/ // / / // // /__
\____//_/ \____/ \__,_/ \__,_/ /_/  /_/ \___/ \___/ /_/ /_/ \__,_//_/ /_//_/ \___/

  REGIONS            CRITICAL  Open SG sg-0db0d4a (caf-bastion-sg) allows SSH from 0.0.0.0/0
  * us-east-1        WARNING   Unattached EBS vol-0abc123 in available state
                     CRITICAL  IAM user admin@corp has no MFA device enabled
  SUMMARY              ...
  3 Critical
  2 Warnings
  5 Total
```

### Dashboard Keybindings

| Key | Action |
|-----|--------|
| `j` / `Down Arrow` | Move to next issue |
| `k` / `Up Arrow` | Move to previous issue |
| `Enter` | Expand issue — show resource ID, scanner, and remediation steps |
| `Esc` | Collapse expanded issue |
| `R` | Re-run the scan |
| `Q` / `Ctrl+C` | Quit the dashboard |

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
- [ ] Custom severity thresholds
- [ ] HTML report export
- [ ] Slack / webhook notifications
- [ ] Cost estimation per issue

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
