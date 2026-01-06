# Factorial HR Setup

This guide explains how to set up and use kspec with Factorial HR for compliance scanning.

## Prerequisites

- A Factorial HR account with admin access
- API access enabled for your Factorial account
- kspec installed on your system

## Authentication

Factorial HR supports two authentication methods:

### API Key (Recommended)

API keys have full access to all Factorial HR data and don't expire. They're ideal for automated scanning and CI/CD pipelines.

1. Log in to Factorial HR as an admin
2. Go to **Settings** > **Integrations** > **API**
3. Click **Generate API Key**
4. Copy and securely store the key

Set the API key as an environment variable:

```bash
export FACTORIAL_API_KEY="your-api-key-here"
```

Or pass it directly to kspec:

```bash
kspec scan factorial --api-key "your-api-key" -f policy.yml
```

### OAuth 2.0 Access Token

OAuth tokens are tied to a specific user and respect their permissions. Use this for user-context operations.

```bash
export FACTORIAL_ACCESS_TOKEN="your-oauth-token"
```

Or pass it directly:

```bash
kspec scan factorial --access-token "your-token" -f policy.yml
```

## Required Permissions

For comprehensive scanning, ensure your API key or OAuth user has access to:

| Resource | Permission | Purpose |
|----------|------------|---------|
| Employees | Read | Scan employee data and status |
| Contracts | Read | Verify contract compliance |
| Teams | Read | Check team structure |
| Time Off | Read | Review leave requests and allowances |
| Attendance | Read | Audit time tracking data |
| Documents | Read | Verify document signatures |
| Locations | Read | Check office configuration |

## Scanning Commands

### Basic Scan

Scan all Factorial HR resources with the default policy:

```bash
kspec scan factorial -f policies/factorial-security.yml
```

### With Explicit Credentials

```bash
kspec scan factorial \
  --api-key "your-api-key" \
  -f policies/factorial-security.yml
```

### Using Environment Variables

```bash
export FACTORIAL_API_KEY="your-api-key"
kspec scan factorial -f policies/factorial-security.yml
```

### Using Provider Aliases

```bash
# These are equivalent:
kspec scan factorial -f policy.yml
kspec scan factorial-hr -f policy.yml
kspec scan hris -f policy.yml
```

## Resources Discovered

kspec discovers and scans the following Factorial HR resources:

### factorial_employee

Employee records with computed compliance fields.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Employee ID |
| `full_name` | string | First and last name |
| `email` | string | Email address |
| `is_active` | bool | Active employment status |
| `has_manager` | bool | Has assigned manager |
| `has_email` | bool | Email is configured |
| `has_team` | bool | Assigned to a team |
| `employment_days` | int | Days since hire date |
| `has_pending_invitation` | bool | Onboarding incomplete |

### factorial_contract

Employment contracts and compensation details.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Contract ID |
| `employee_id` | int | Associated employee |
| `job_title` | string | Position title |
| `is_active` | bool | Contract is active |
| `is_current` | bool | Effective date has passed |
| `is_permanent` | bool | No end date |
| `has_job_title` | bool | Job title defined |
| `has_working_hours` | bool | Hours are configured |
| `has_salary` | bool | Salary is configured |

### factorial_team

Team structure and membership.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Team ID |
| `name` | string | Team name |
| `member_count` | int | Number of members |
| `has_lead` | bool | Has designated lead |
| `has_members` | bool | Has at least one member |

### factorial_leave

Time off requests and approvals.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Leave request ID |
| `employee_id` | int | Requesting employee |
| `status` | string | pending/approved/rejected |
| `is_approved` | bool | Request approved |
| `is_pending` | bool | Awaiting approval |
| `is_stale` | bool | Pending > 7 days |
| `leave_days` | float | Duration in days |

### factorial_allowance

Leave entitlements per employee.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Allowance ID |
| `employee_id` | int | Employee |
| `allowance` | float | Total days allowed |
| `remaining` | float | Days remaining |
| `has_allowance` | bool | Allowance configured |
| `is_exhausted` | bool | No days remaining |

### factorial_document

HR documents and signature status.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Document ID |
| `name` | string | Document name |
| `employee_id` | int | Associated employee |
| `is_signed` | bool | Document is signed |
| `is_pending` | bool | Awaiting signature |
| `needs_signature` | bool | Signature required but missing |
| `is_nda` | bool | Non-disclosure agreement |
| `is_contract` | bool | Employment contract |
| `is_policy` | bool | Company policy |

### factorial_location

Office locations and configuration.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Location ID |
| `name` | string | Location name |
| `country` | string | Country code |
| `has_address` | bool | Address configured |
| `has_timezone` | bool | Timezone configured |
| `employee_count` | int | Employees at location |

## Compliance Use Cases

### SOC 2

- Verify employee access reviews (managers assigned)
- Audit termination procedures (inactive employees)
- Check document acknowledgments (policies signed)

### ISO 27001

- Organizational structure validation (teams/roles)
- Access control verification (job titles defined)
- Information security training (training documents signed)

### GDPR

- Employee data completeness (emails, addresses)
- Right to access readiness (document availability)
- Data subject records (employee profiles complete)

### NIS2

- Workforce management (active employee tracking)
- Incident response teams (team structure defined)
- Access control policies (manager approvals)

## Example Policy

See `policies/factorial-security.yml` for a complete example policy with checks for:

- Employee management compliance
- Contract compliance
- Team organization
- Time off management
- Document management
- Location configuration

## Troubleshooting

### Authentication Errors

```
factorial: no credentials provided
```

**Solution:** Set `FACTORIAL_API_KEY` or `FACTORIAL_ACCESS_TOKEN` environment variable, or use `--api-key` flag.

### API Rate Limits

Factorial API has rate limits:
- POST requests: 200/minute on v2 endpoints
- POST requests: 100/minute on v1 endpoints

If you hit rate limits, wait a few minutes before retrying.

### Empty Results

If scans return no resources:
1. Verify API key has appropriate permissions
2. Check if your Factorial account has data
3. Ensure you're connecting to the correct account

### Connection Errors

```
factorial: failed to connect
```

**Solutions:**
- Verify API key is valid and not expired
- Check network connectivity to api.factorialhr.com
- Ensure your IP is not blocked by Factorial

## CI/CD Integration

### GitHub Actions

```yaml
name: HR Compliance Scan
on:
  schedule:
    - cron: '0 9 * * 1'  # Weekly on Monday

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install kspec
        run: |
          curl -sSL https://get.kspec.dev | bash

      - name: Run Factorial scan
        env:
          FACTORIAL_API_KEY: ${{ secrets.FACTORIAL_API_KEY }}
        run: |
          kspec scan factorial -f policies/factorial-security.yml
```

### GitLab CI

```yaml
factorial-scan:
  stage: compliance
  script:
    - kspec scan factorial -f policies/factorial-security.yml
  variables:
    FACTORIAL_API_KEY: $FACTORIAL_API_KEY
  only:
    - schedules
```

## API Reference

The Factorial provider uses API version `2025-07-01`. For more information about the Factorial API, see:

- [Factorial API Documentation](https://apidoc.factorialhr.com/)
- [Factorial API Reference](https://apidoc.factorialhr.com/reference)
- [Factorial Help Center](https://help.factorialhr.com/)
