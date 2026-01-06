# Dynamic Sub-Resource System

## Overview

The dynamic sub-resource system allows resource providers to automatically register dependent resources at runtime. This enables hierarchical resource relationships without hardcoding them in the provider registry.

## How It Works

### 1. Core Interface: `SubResourceProvider`

Resources can optionally implement the `SubResourceProvider` interface:

```go
type SubResourceProvider interface {
    ResourceSpec
    // SubResources returns additional resource specs that should be registered
    SubResources() []ResourceSpec
}
```

### 2. Automatic Discovery

When a provider is initialized, the registry automatically:
1. Registers all primary resources from `Connection.Resources()`
2. Checks each resource for `SubResourceProvider` implementation
3. Calls `SubResources()` on implementing resources
4. Registers all returned sub-resources

### 3. Example: SQL Server → Databases

```go
// SQL Server is a primary resource
type SqlServerResource struct {
    credential     azcore.TokenCredential
    subscriptionID string
}

// Implements SubResourceProvider
func (r *SqlServerResource) SubResources() []core.ResourceSpec {
    return []core.ResourceSpec{
        &SqlDatabaseResource{
            credential:     r.credential,
            subscriptionID: r.subscriptionID,
        },
    }
}
```

Now when Azure provider connects:
- `azure_sql_server` is registered (primary)
- `azure_sql_database` is automatically discovered and registered (sub-resource)

## Benefits

### ✅ No Manual Registration
Sub-resources don't need to be listed in `provider.go`. They're discovered automatically.

### ✅ Hierarchical Context
Sub-resources inherit credentials and context from their parent:
```go
&SqlDatabaseResource{
    credential:     r.credential,      // From parent
    subscriptionID: r.subscriptionID,  // From parent
}
```

### ✅ Provider Isolation
Each provider manages its own resource hierarchy independently.

### ✅ Lazy Loading
Sub-resources are only registered when their parent provider is loaded.

## Use Cases

### Azure
- `azure_sql_server` → `azure_sql_database`
- `azure_storage_account` → `azure_blob_container`
- `azure_keyvault_vault` → `azure_keyvault_key`, `azure_keyvault_secret`
- `azure_app_service` → `azure_app_service_slot`

### GitHub
- `github_repo` → `github_workflow`, `github_environment`
- `github_org` → `github_team`, `github_webhook`

### AWS (Future)
- `aws_rds_cluster` → `aws_rds_instance`
- `aws_ecs_cluster` → `aws_ecs_service` → `aws_ecs_task`

## Implementation Guide

### Step 1: Create Sub-Resource

```go
type MySubResource struct {
    // Inherit from parent
    parentCredential azcore.TokenCredential
    parentID         string
}

func (r *MySubResource) Name() string {
    return "my_sub_resource"
}

func (r *MySubResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
    // Fetch sub-resources using parent context
    // ...
}
```

### Step 2: Implement SubResourceProvider on Parent

```go
func (r *MyParentResource) SubResources() []core.ResourceSpec {
    return []core.ResourceSpec{
        &MySubResource{
            parentCredential: r.credential,
            parentID:         r.id,
        },
        // Add more sub-resources as needed
    }
}
```

### Step 3: Use in Policies

```yaml
queries:
  - uid: check-sub-resource
    title: Check Sub Resource
    resource: my_sub_resource  # Automatically available!
    query: |
      resource.some_property == true
```

## Advanced: Multi-Level Hierarchies

Sub-resources can themselves be `SubResourceProvider`:

```go
type Level1Resource struct {}

func (r *Level1Resource) SubResources() []core.ResourceSpec {
    return []core.ResourceSpec{
        &Level2Resource{},  // Level 2
    }
}

type Level2Resource struct {}

func (r *Level2Resource) SubResources() []core.ResourceSpec {
    return []core.ResourceSpec{
        &Level3Resource{},  // Level 3
    }
}
```

The registry will recursively discover all levels:
- `level1_resource`
- `level2_resource`
- `level3_resource`

## Performance Considerations

### Lazy Instantiation
Sub-resources are only created during provider initialization, not during every Fetch().

### Shared Credentials
Credentials are shared from parent to child, avoiding redundant authentication.

### Conditional Registration
Sub-resources can be conditionally registered based on parent state:

```go
func (r *SqlServerResource) SubResources() []core.ResourceSpec {
    subResources := []core.ResourceSpec{}
    
    // Only register if certain conditions are met
    if r.supportsDatabases {
        subResources = append(subResources, &SqlDatabaseResource{...})
    }
    
    return subResources
}
```

## Testing

```go
// Test that sub-resources are registered
func TestSubResourceRegistration(t *testing.T) {
    registry, err := provider.InitProvider(ctx, "azure", config)
    require.NoError(t, err)
    
    // Check primary resource
    assert.Contains(t, registry, "azure_sql_server")
    
    // Check sub-resource is automatically registered
    assert.Contains(t, registry, "azure_sql_database")
}
```

## Migration Path

Existing resources continue to work without changes. To add sub-resources:

1. Create new sub-resource types
2. Implement `SubResources()` on parent
3. Sub-resources are immediately available in policies

No changes needed to:
- Provider registry
- CLI code  
- Policy engine
- Existing policies
