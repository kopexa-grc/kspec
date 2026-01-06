# CEL Expressions

kspec uses [CEL (Common Expression Language)](https://github.com/google/cel-spec) for policy queries. CEL is a fast, portable expression language designed for security policies.

## Overview

Every policy query is a CEL expression that evaluates to `true` (pass) or `false` (fail):

```yaml
query: resource.enabled == true
```

The `resource` variable contains the current resource being evaluated.

## Basic Operations

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `resource.status == "active"` |
| `!=` | Not equal | `resource.public != true` |
| `<` | Less than | `resource.count < 10` |
| `<=` | Less or equal | `resource.score <= 50` |
| `>` | Greater than | `resource.size > 0` |
| `>=` | Greater or equal | `resource.version >= "1.2"` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | AND | `a == true && b == true` |
| `\|\|` | OR | `a == true \|\| b == true` |
| `!` | NOT | `!resource.public` |

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `resource.a + resource.b` |
| `-` | Subtraction | `resource.total - resource.used` |
| `*` | Multiplication | `resource.count * 2` |
| `/` | Division | `resource.total / resource.count` |
| `%` | Modulo | `resource.id % 10` |

## Field Access

### Direct Access

```yaml
# Access field
query: resource.name

# Nested access
query: resource.properties.encryption.enabled

# With default for missing field
query: resource.tags.getOrDefault("env", "unknown")
```

### Field Existence

```yaml
# Check if field exists
query: has(resource.encryption)

# Check nested field
query: has(resource.properties.networkAcls)

# Combine with value check
query: |
  has(resource.encryption) &&
  resource.encryption.enabled == true
```

## String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `contains(s)` | Contains substring | `resource.name.contains("prod")` |
| `startsWith(s)` | Starts with prefix | `resource.name.startsWith("app-")` |
| `endsWith(s)` | Ends with suffix | `resource.name.endsWith("-dev")` |
| `matches(re)` | Regex match | `resource.name.matches("^[a-z]+$")` |
| `size()` | String length | `resource.name.size() > 0` |
| `lowerAscii()` | Lowercase | `resource.type.lowerAscii() == "blob"` |
| `upperAscii()` | Uppercase | `resource.code.upperAscii() == "US"` |

### String Examples

```yaml
# Check naming convention
query: resource.name.matches("^prod-[a-z]+-[0-9]+$")

# Case-insensitive comparison
query: resource.type.lowerAscii() == "storage"

# Check for substring
query: resource.description.contains("encrypted")

# Check prefix
query: |
  resource.name.startsWith("secure-") ||
  resource.name.startsWith("prod-")
```

## List Operations

### Basic List Functions

| Function | Description | Example |
|----------|-------------|---------|
| `size()` | List length | `size(resource.rules) > 0` |
| `all(x, expr)` | All match | `resource.items.all(i, i.valid)` |
| `exists(x, expr)` | Any match | `resource.items.exists(i, i.type == "A")` |
| `exists_one(x, expr)` | Exactly one | `resource.items.exists_one(i, i.primary)` |
| `filter(x, expr)` | Filter list | `resource.items.filter(i, i.active)` |
| `map(x, expr)` | Transform | `resource.items.map(i, i.name)` |

### List Examples

```yaml
# Check list has items
query: size(resource.members) > 0

# All items pass condition
query: |
  resource.rules.all(rule,
    rule.action == "allow" || rule.source != "0.0.0.0/0"
  )

# At least one item matches
query: |
  resource.tags.exists(tag,
    tag.key == "environment"
  )

# No items match (dangerous condition)
query: |
  !resource.rules.exists(rule,
    rule.port == "22" &&
    rule.source == "0.0.0.0/0"
  )

# Check specific values exist
query: |
  resource.versions.exists(v, v == "TLSv1.2") &&
  resource.versions.exists(v, v == "TLSv1.3")
```

## Map Operations

### Map Functions

| Function | Description | Example |
|----------|-------------|---------|
| `has(key)` | Key exists | `has(resource.tags.env)` |
| `size()` | Map size | `size(resource.labels) > 0` |

### Map Examples

```yaml
# Check tag exists
query: has(resource.tags.environment)

# Check tag value
query: |
  has(resource.tags.environment) &&
  resource.tags.environment == "production"

# Check map is not empty
query: size(resource.labels) > 0
```

## Conditional Logic

### Ternary Operator

```yaml
# condition ? true_value : false_value
query: |
  resource.type == "public" ?
    resource.encrypted == true :
    true
```

### Complex Conditions

```yaml
# Multiple conditions with fallback
query: |
  (resource.type == "critical" && resource.backup_enabled == true) ||
  (resource.type == "standard" && has(resource.backup_policy)) ||
  resource.type == "temporary"
```

## Type Handling

### Type Functions

| Function | Description | Example |
|----------|-------------|---------|
| `type()` | Get type | `type(resource.value)` |
| `int()` | Convert to int | `int(resource.count)` |
| `double()` | Convert to double | `double(resource.score)` |
| `string()` | Convert to string | `string(resource.id)` |

### Type Checking

```yaml
# Check type before comparison
query: |
  type(resource.count) == int &&
  resource.count >= 2
```

## Common Patterns

### Safe Field Access

Always check field existence before accessing nested fields:

```yaml
# Bad: May fail if properties doesn't exist
query: resource.properties.enabled == true

# Good: Check existence first
query: |
  has(resource.properties) &&
  has(resource.properties.enabled) &&
  resource.properties.enabled == true
```

### Optional Fields

Handle optional fields gracefully:

```yaml
# Pass if field missing OR has correct value
query: |
  !has(resource.public_access) ||
  resource.public_access == false
```

### Skip Non-Applicable Resources

Skip resources that don't apply to the check:

```yaml
# Only check default branches
query: |
  resource.is_default == false ||
  resource.protected == true

# Only check production resources
query: |
  !resource.name.contains("prod") ||
  resource.encrypted == true
```

### Check Multiple Conditions

```yaml
# All security settings enabled
query: |
  resource.encryption.enabled == true &&
  resource.logging.enabled == true &&
  resource.monitoring.enabled == true &&
  !resource.public_access
```

### Validate Lists

```yaml
# No dangerous rules
query: |
  !resource.firewall_rules.exists(rule,
    rule.action == "allow" &&
    rule.direction == "inbound" &&
    (rule.source == "*" || rule.source == "0.0.0.0/0") &&
    (rule.port == "*" || rule.port == "22" || rule.port == "3389")
  )

# At least one admin
query: |
  resource.members.exists(m,
    m.role == "admin" || m.role == "owner"
  ) &&
  size(resource.members.filter(m, m.role == "admin")) <= 4
```

### Check File Existence

```yaml
# Has security configuration file
query: |
  resource.files.exists(f,
    f.path == ".github/dependabot.yml" ||
    f.path == ".github/dependabot.yaml" ||
    f.path == "dependabot.yml"
  )
```

## Debugging Tips

### Simplify Expressions

Break complex queries into parts:

```yaml
# Instead of one complex query
query: |
  has(resource.a) && resource.a.b == true && has(resource.c) && resource.c.d.exists(x, x.e == "f")

# Test each part
query: has(resource.a)
query: resource.a.b == true
query: has(resource.c)
query: resource.c.d.exists(x, x.e == "f")
```

### Use Meaningful Variables

```yaml
# Clear variable names in list operations
query: |
  resource.security_rules.all(rule,
    rule.direction != "inbound" ||
    rule.source_address != "0.0.0.0/0"
  )
```

## Performance Tips

1. **Check existence first**: `has()` is fast
2. **Short-circuit evaluation**: Put cheap checks first
3. **Avoid complex regex**: Simple string ops are faster
4. **Limit list iterations**: Filter early when possible

```yaml
# Optimized: cheap check first
query: |
  resource.type != "critical" ||
  (has(resource.encryption) && resource.encryption.enabled == true)
```

## Further Reading

- [CEL Specification](https://github.com/google/cel-spec)
- [CEL Go Implementation](https://github.com/google/cel-go)
- [Writing Policies](../policies.md)
