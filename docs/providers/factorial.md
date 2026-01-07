# Factorial HR Provider

The Factorial HR provider scans your Factorial HR account for compliance with HR policies, including employee data, contracts, attendance, and documents.

## Overview

Use the Factorial HR provider to validate:
- Employee information and status
- Contract compliance (types, dates, terms)
- Team structures and memberships
- Time off and leave management
- Document management and compliance
- Location and workplace settings

## Quick Start

```bash
# Set your API key
export FACTORIAL_API_KEY="your-api-key"

# Scan with Factorial policies
kspec scan factorial tenant -f policies/factorial-compliance.yml
```

## Prerequisites

- Factorial HR account with API access
- API Key or OAuth Access Token

## Authentication

### API Key Authentication

1. Log in to your Factorial HR admin account
2. Navigate to **Settings** > **Integrations** > **API**
3. Generate a new API key
4. Copy the key securely

**Environment Variable (Recommended):**
```bash
export FACTORIAL_API_KEY="your-api-key"
kspec scan factorial tenant -f policy.yml
```

### OAuth Access Token

For OAuth-based authentication:

```bash
export FACTORIAL_ACCESS_TOKEN="your-oauth-token"
kspec scan factorial tenant -f policy.yml
```

### Configuration Options

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `api_key` | `FACTORIAL_API_KEY` | API key for authentication |
| `access_token` | `FACTORIAL_ACCESS_TOKEN` | OAuth access token |
| `base_url` | `FACTORIAL_BASE_URL` | API base URL (default: `https://api.factorialhr.com`) |
| `api_version` | `FACTORIAL_API_VERSION` | API version (default: `2025-07-01`) |

---

## Resources

The Factorial HR provider discovers the following resources:

### Core HR

| Resource | Description |
|----------|-------------|
| `factorial_employee` | Employee profiles and information |
| `factorial_contract` | Employment contracts |

### Organization

| Resource | Description |
|----------|-------------|
| `factorial_team` | Teams and departments |
| `factorial_team_membership` | Team member assignments |
| `factorial_location` | Office locations and workplaces |

### Time & Attendance

| Resource | Description |
|----------|-------------|
| `factorial_shift` | Work shifts and schedules |
| `factorial_leave` | Time off requests and leaves |
| `factorial_allowance` | Leave allowances and balances |

### Documents

| Resource | Description |
|----------|-------------|
| `factorial_document` | HR documents |
| `factorial_folder` | Document folders |

---

## Resource Fields

### factorial_employee

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Employee ID |
| `first_name` | string | First name |
| `last_name` | string | Last name |
| `email` | string | Work email |
| `status` | string | Employment status |
| `hire_date` | string | Date of hire |
| `termination_date` | string | Termination date (if applicable) |
| `manager_id` | string | Manager's employee ID |
| `team_ids` | array | Associated team IDs |
| `location_id` | string | Work location ID |

### factorial_contract

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Contract ID |
| `employee_id` | string | Associated employee ID |
| `contract_type` | string | Contract type (permanent, temporary, etc.) |
| `start_date` | string | Contract start date |
| `end_date` | string | Contract end date |
| `salary` | number | Salary amount |
| `working_hours` | number | Weekly working hours |
| `probation_end_date` | string | Probation period end date |

---

## Example Policies

### Contract Compliance

```yaml
policies:
  - uid: factorial-compliance
    name: Factorial HR Compliance
    version: 1.0.0
    require:
      - provider: factorial
    groups:
      - title: Contract Compliance
        checks:
          - uid: contract-has-end-date
          - uid: employee-has-manager

queries:
  - uid: contract-has-end-date
    title: Ensure temporary contracts have end dates
    resource: factorial_contract
    impact: 70
    query: |
      resource.contract_type != "temporary" ||
      (has(resource.end_date) && resource.end_date != "")
    docs:
      desc: Temporary contracts must have a defined end date for compliance.
      remediation: Update the contract to include an end date.

  - uid: employee-has-manager
    title: Ensure employees have assigned managers
    resource: factorial_employee
    impact: 50
    query: |
      resource.status != "active" ||
      (has(resource.manager_id) && resource.manager_id != "")
    docs:
      desc: Active employees should have an assigned manager for proper reporting structure.
      remediation: Assign a manager to the employee in Factorial HR.
```

### Document Compliance

```yaml
queries:
  - uid: required-documents
    title: Ensure required HR documents exist
    resource: factorial_document
    impact: 80
    query: |
      resource.status == "signed" || resource.status == "completed"
    docs:
      desc: HR documents should be properly signed and completed.
      remediation: Follow up with the employee to complete and sign required documents.
```

---

## Troubleshooting

### Authentication Errors

```
factorial: no credentials provided
```

**Solutions:**
- Set `FACTORIAL_API_KEY` or `FACTORIAL_ACCESS_TOKEN`
- Verify the API key is valid and not expired
- Check API key permissions in Factorial HR settings

### API Errors

```
factorial: API error 401: Unauthorized
```

**Solutions:**
- Regenerate your API key in Factorial HR
- Verify you're using the correct authentication method
- Check if your account has API access enabled

### Rate Limiting

```
factorial: API error 429: Too Many Requests
```

**Solutions:**
- Reduce scan frequency
- Contact Factorial HR support for rate limit increase
