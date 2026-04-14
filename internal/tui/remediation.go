package tui

import (
	"github.com/cloudmechanic-cli/cloudmechanic/internal/scanner"
)

// Remediation holds the Terraform fix for a specific scanner type.
type Remediation struct {
	Title         string
	Description   string
	TerraformCode string
}

// noRemediation is the fallback when no specific snippet is mapped.
var noRemediation = &Remediation{
	Title:       "Manual Review Required",
	Description: "No automated Terraform snippet is available for this issue type. Refer to the AWS documentation for manual remediation steps.",
	TerraformCode: `# No automated Terraform snippet for this issue.
#
# Refer to the AWS documentation for manual remediation steps:
#   https://docs.aws.amazon.com/`,
}

// remediationMap maps scanner Name() → Remediation snippet.
var remediationMap = map[string]*Remediation{

	// ── S3 ────────────────────────────────────────────────────────────────────

	"S3 Buckets Without Versioning": {
		Title:       "Enable S3 Bucket Versioning",
		Description: "Versioning protects against accidental deletion and overwrites by retaining every object version. Required for cross-region replication and S3 Object Lock.",
		TerraformCode: `resource "aws_s3_bucket_versioning" "fix" {
  bucket = "<YOUR_BUCKET_NAME>"

  versioning_configuration {
    status = "Enabled"
  }
}`,
	},

	"Public S3 Buckets": {
		Title:       "Enable S3 Public Access Block",
		Description: "Enabling all four Public Access Block settings prevents any future ACL or bucket policy from making the bucket or its objects publicly accessible.",
		TerraformCode: `resource "aws_s3_bucket_public_access_block" "fix" {
  bucket = "<YOUR_BUCKET_NAME>"

  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = true
  restrict_public_buckets = true
}`,
	},

	"S3 Buckets Without Encryption": {
		Title:       "Enable S3 Default Server-Side Encryption",
		Description: "Configuring SSE-KMS as the bucket default encrypts every new object at rest automatically, satisfying most compliance requirements (PCI-DSS, HIPAA, SOC 2).",
		TerraformCode: `resource "aws_s3_bucket_server_side_encryption_configuration" "fix" {
  bucket = "<YOUR_BUCKET_NAME>"

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
      # Optionally specify your own KMS key:
      # kms_master_key_id = aws_kms_key.s3.arn
    }

    # Reduces SSE-KMS API costs by caching the data key.
    bucket_key_enabled = true
  }
}`,
	},

	// ── VPC ───────────────────────────────────────────────────────────────────

	"VPCs Without Flow Logs": {
		Title:       "Enable VPC Flow Logs to CloudWatch",
		Description: "Flow Logs capture metadata for every IP packet crossing your VPC (accepted and rejected). Essential for security forensics, intrusion detection, and compliance audits.",
		TerraformCode: `# 1. IAM role that VPC Flow Logs will assume.
resource "aws_iam_role" "flow_logs" {
  name = "vpc-flow-logs-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "vpc-flow-logs.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "flow_logs" {
  name = "vpc-flow-logs-policy"
  role = aws_iam_role.flow_logs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams"
      ]
      Resource = "*"
    }]
  })
}

# 2. CloudWatch Log Group to receive the flow log records.
resource "aws_cloudwatch_log_group" "flow_logs" {
  name              = "/aws/vpc/flow-logs/<YOUR_VPC_ID>"
  retention_in_days = 30
}

# 3. Enable Flow Logs on your VPC.
resource "aws_flow_log" "fix" {
  iam_role_arn    = aws_iam_role.flow_logs.arn
  log_destination = aws_cloudwatch_log_group.flow_logs.arn
  traffic_type    = "ALL"
  vpc_id          = "<YOUR_VPC_ID>"
}`,
	},

	"Unused NAT Gateways": {
		Title:       "Remove Unused NAT Gateway",
		Description: "An idle NAT Gateway costs ~$32/month in hourly fees plus data-processing charges. Delete it if no private subnets need outbound internet access.",
		TerraformCode: `# Option A – The NAT Gateway was created by a Terraform resource block.
# Delete the block and apply:
#
#   resource "aws_nat_gateway" "this" { ... }   ← remove this
#   resource "aws_eip" "nat"          { ... }   ← remove this too
#
# Then run: terraform apply

# Option B – Using the official VPC module.
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  # Disable NAT Gateways entirely:
  enable_nat_gateway = false

  # Or keep one shared NAT Gateway (reduces cost, not HA):
  # enable_nat_gateway  = true
  # single_nat_gateway  = true

  # ... rest of your VPC config
}`,
	},

	// ── EC2 ───────────────────────────────────────────────────────────────────

	"Open Security Groups (SSH)": {
		Title:       "Restrict SSH Ingress to a Trusted CIDR",
		Description: "Opening port 22 to 0.0.0.0/0 exposes your instances to brute-force attacks from the entire internet. Restrict access to a known CIDR or use AWS Systems Manager Session Manager instead.",
		TerraformCode: `# Replace the wide-open rule with a trusted CIDR.
resource "aws_security_group_rule" "ssh_restricted" {
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  security_group_id = "<YOUR_SECURITY_GROUP_ID>"

  # Replace with your office/VPN CIDR or a bastion security group:
  cidr_blocks = ["10.0.0.0/8"]
  description = "SSH — restricted to trusted network"
}

# Revoke the existing open rule via the AWS CLI:
# aws ec2 revoke-security-group-ingress \
#   --group-id <YOUR_SECURITY_GROUP_ID> \
#   --protocol tcp --port 22 --cidr 0.0.0.0/0`,
	},

	"Unused Elastic IPs": {
		Title:       "Release or Associate Unused Elastic IP",
		Description: "AWS charges ~$0.005/hour for every Elastic IP that is not associated with a running instance. Release unused EIPs or associate them immediately.",
		TerraformCode: `# Option A – Associate the EIP with a running EC2 instance.
resource "aws_eip_association" "fix" {
  instance_id   = "<YOUR_INSTANCE_ID>"
  allocation_id = "<YOUR_EIP_ALLOCATION_ID>"
}

# Option B – If the EIP was created by Terraform, remove its
# resource block to release it on the next apply:
#
#   resource "aws_eip" "this" { ... }   ← delete this block
#
# Or release via the AWS CLI:
# aws ec2 release-address --allocation-id <YOUR_EIP_ALLOCATION_ID>`,
	},

	"Unattached EBS Volumes": {
		Title:       "Attach or Delete Unattached EBS Volume",
		Description: "Unattached EBS volumes incur storage costs indefinitely. Attach them to a running instance, or take a final snapshot and delete.",
		TerraformCode: `# Option A – Attach the volume to an EC2 instance.
resource "aws_volume_attachment" "fix" {
  device_name = "/dev/xvdf"
  volume_id   = "<YOUR_VOLUME_ID>"
  instance_id = "<YOUR_INSTANCE_ID>"
}

# Option B – Snapshot first, then delete (AWS CLI).
# aws ec2 create-snapshot \
#   --volume-id <YOUR_VOLUME_ID> \
#   --description "Final snapshot before deletion"
#
# aws ec2 delete-volume --volume-id <YOUR_VOLUME_ID>`,
	},

	"Old EBS Snapshots (>90 days)": {
		Title:       "Automate EBS Snapshot Lifecycle with DLM",
		Description: "AWS Data Lifecycle Manager (DLM) creates snapshots on a schedule and automatically expires old ones, preventing unbounded snapshot accumulation.",
		TerraformCode: `resource "aws_iam_role" "dlm" {
  name = "dlm-lifecycle-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "dlm.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "dlm" {
  role       = aws_iam_role.dlm.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSDataLifecycleManagerServiceRole"
}

resource "aws_dlm_lifecycle_policy" "fix" {
  description        = "Daily EBS snapshots — 30-day retention"
  execution_role_arn = aws_iam_role.dlm.arn
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "Daily snapshots"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["23:45"]
      }

      retain_rule {
        count = 30
      }

      copy_tags = true
    }

    # Only volumes with this tag will be snapshotted:
    target_tags = {
      Backup = "true"
    }
  }
}`,
	},

	// ── RDS ───────────────────────────────────────────────────────────────────

	"Idle RDS Instances": {
		Title:       "Stop or Delete Idle RDS Instance",
		Description: "RDS instances accrue hourly instance charges even when idle. Stop the instance for temporary relief (AWS auto-starts after 7 days), or delete if no longer needed.",
		TerraformCode: `# Option A – Stop temporarily via the AWS CLI:
# aws rds stop-db-instance --db-instance-identifier <YOUR_DB_ID>
# Note: AWS will automatically restart it after 7 days.

# Option B – Delete with a final snapshot (Terraform).
resource "aws_db_instance" "example" {
  identifier = "<YOUR_DB_ID>"

  # Set these before running terraform destroy:
  skip_final_snapshot       = false
  final_snapshot_identifier = "<YOUR_DB_ID>-final"

  # Then destroy only this resource:
  # terraform destroy -target=aws_db_instance.example
}

# Option C – Use scheduled start/stop for dev/test instances
# (EventBridge + Lambda) so they only run during business hours.`,
	},

	// ── DynamoDB ─────────────────────────────────────────────────────────────

	"DynamoDB Without Backups": {
		Title:       "Enable DynamoDB Point-in-Time Recovery (PITR)",
		Description: "PITR provides continuous, automatic backups with a rolling 35-day recovery window at no additional upfront cost. Essential for production tables.",
		TerraformCode: `resource "aws_dynamodb_table" "fix" {
  name     = "<YOUR_TABLE_NAME>"
  hash_key = "<YOUR_HASH_KEY>"

  # ... your existing table configuration ...

  # Add this block to enable PITR:
  point_in_time_recovery {
    enabled = true
  }
}`,
	},

	"DynamoDB Provisioned Capacity": {
		Title:       "Switch DynamoDB to On-Demand Billing",
		Description: "On-demand mode charges only for the read/write request units you actually consume. Ideal for tables with unpredictable or low traffic that leave provisioned capacity idle.",
		TerraformCode: `resource "aws_dynamodb_table" "fix" {
  name         = "<YOUR_TABLE_NAME>"
  billing_mode = "PAY_PER_REQUEST"  # was: "PROVISIONED"

  # Remove these attributes when switching away from PROVISIONED:
  # read_capacity  = 5
  # write_capacity = 5

  hash_key = "<YOUR_HASH_KEY>"

  attribute {
    name = "<YOUR_HASH_KEY>"
    type = "S"
  }

  # Re-add any GSIs, LSIs, TTL, tags, etc. below
}`,
	},

	// ── IAM ───────────────────────────────────────────────────────────────────

	"IAM Users Without MFA": {
		Title:       "Enforce MFA via IAM Deny Policy",
		Description: "Attach a policy that denies every AWS action except MFA self-management unless the caller authenticated with a second factor. This is an immediate hard-stop control.",
		TerraformCode: `resource "aws_iam_policy" "require_mfa" {
  name        = "RequireMFA"
  description = "Deny all API actions when MFA is not present"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyWithoutMFA"
        Effect = "Deny"
        NotAction = [
          "iam:CreateVirtualMFADevice",
          "iam:EnableMFADevice",
          "iam:GetUser",
          "iam:ListMFADevices",
          "iam:ListVirtualMFADevices",
          "iam:ResyncMFADevice",
          "sts:GetSessionToken"
        ]
        Resource = "*"
        Condition = {
          BoolIfExists = {
            "aws:MultiFactorAuthPresent" = "false"
          }
        }
      }
    ]
  })
}

# Attach the policy to the user (or a group they belong to):
resource "aws_iam_user_policy_attachment" "mfa" {
  user       = "<YOUR_IAM_USERNAME>"
  policy_arn = aws_iam_policy.require_mfa.arn
}`,
	},

	// ── Lambda ────────────────────────────────────────────────────────────────

	"Lambda Deprecated Runtimes": {
		Title:       "Update Lambda Function to a Supported Runtime",
		Description: "Deprecated runtimes no longer receive security patches. Update to a current runtime and redeploy your deployment package.",
		TerraformCode: `resource "aws_lambda_function" "fix" {
  function_name = "<YOUR_FUNCTION_NAME>"
  filename      = "<YOUR_DEPLOYMENT_PACKAGE.zip>"
  handler       = "index.handler"
  role          = aws_iam_role.lambda_exec.arn

  # Replace with a currently supported runtime:
  runtime = "python3.12"

  # Supported runtimes (as of 2025):
  #   Python  : python3.12  python3.11  python3.10
  #   Node.js : nodejs22.x  nodejs20.x
  #   Java    : java21      java17      java11
  #   .NET    : dotnet8
  #   Go      : provided.al2023  (custom runtime)

  source_code_hash = filebase64sha256("<YOUR_DEPLOYMENT_PACKAGE.zip>")
}`,
	},

	"Lambda Public Function URLs": {
		Title:       "Require IAM Auth on Lambda Function URL",
		Description: "A Function URL with AuthType NONE is publicly accessible by anyone on the internet. Switch to AWS_IAM auth or front the function with API Gateway for granular access control.",
		TerraformCode: `# Option A – Require AWS IAM SigV4 authentication.
resource "aws_lambda_function_url" "fix" {
  function_name      = "<YOUR_FUNCTION_NAME>"
  authorization_type = "AWS_IAM"  # was: "NONE"
}

# Grant a specific IAM role permission to call the URL:
resource "aws_lambda_permission" "allow_caller" {
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = "<YOUR_FUNCTION_NAME>"
  principal              = "<CALLER_ROLE_ARN>"
  function_url_auth_type = "AWS_IAM"
}

# Option B – Remove the Function URL entirely and use
# API Gateway for full request/response control:
#
#   resource "aws_lambda_function_url" "this" { ... }  ← delete
#   resource "aws_api_gateway_rest_api" "api"  { ... }  ← add`,
	},
}

// GetRemediation returns the Terraform remediation for the given issue.
// Falls back to a generic snippet when no specific mapping exists.
func GetRemediation(issue scanner.Issue) *Remediation {
	if rem, ok := remediationMap[issue.Scanner]; ok {
		return rem
	}
	return noRemediation
}
