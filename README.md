# Starter CLI 🚀

A powerful Go code generation tool that automatically generates entities, resources, modules, builders, and routes from your database migrations.

## Features

- 🗄️ **Entity Generation** - Generate Go entities from SQL migration files
- 📦 **Resource Generation** - Create DTOs (Data Transfer Objects) for API layers
- 🏗️ **Module Generation** - Generate handlers, services, and repositories
- 🔧 **Builder Generation** - Automatic dependency injection setup
- 🛣️ **Routes Generation** - REST API routes with middleware support
- 🔐 **Auto Permissions** - Automatic permission constant generation
- 💾 **Auto Cache Keys** - Automatic Redis cache key generation
- 🎨 **Custom Templates** - Fully customizable template system
- 🔄 **Incremental Updates** - Smart updates to existing files

## Installation

```bash
go install github.com/rifqiakrm/starter-cli/cmd/starter-cli@latest
```

This will install the `starter-cli` binary to your `$GOPATH/bin`.

## Quick Start

### Generate Complete Stack
```bash
# Generate entity, resource, and module components
starter-cli all --schema=auth --table=users --version=v1
```

### Generate Individual Components
```bash
# Generate only entity
starter-cli entity --schema=auth --table=users

# Generate only resource DTOs
starter-cli resource --schema=auth --table=users

# Generate module components
starter-cli module --schema=auth --table=users --version=v1 --parts=handler,service,repository
```

### Module Builder
```bash
# Create complete new module
starter-cli builder --module=auth --tables=users,roles,permissions --version=v1 --new-module

# Add tables to existing module
starter-cli builder --module=auth --tables=organizations --version=v1

# Preview changes (dry run)
starter-cli builder --module=auth --tables=organizations --dry-run
```

## Commands

| Command | Description |
|---------|-------------|
| `all` | Generate entity, resource, and module components |
| `entity` | Generate entity from database table |
| `resource` | Generate resource DTOs from database table |
| `module` | Generate module components (handler, service, repository) |
| `builder` | Generate builder and routes for modules |
| `init` | Initialize template directory for customization |
| `help` | Show usage information |
| `version` | Show version information |

## Command Examples

### Basic Usage
```bash
# Generate complete stack for users table
starter-cli all --schema=auth --table=users --version=v1

# Generate only entity
starter-cli entity --schema=inventory --table=products

# Generate specific module parts
starter-cli module --schema=inventory --table=categories --version=v1 --parts=handler.creator,service.finder
```

### Module Parts Syntax
```bash
--parts=handler                    # All handler actions
--parts=handler.creator            # Only creator handler
--parts=handler.finder,service     # Finder handler + all services
--parts=repository.creator,repository.updater  # Specific repository actions
```

### Template Customization
```bash
# Initialize template directory
starter-cli init

# Use custom templates
starter-cli all --schema=auth --table=users --version=v1 --template-dir=./templates
```

## Flags

### Common Flags
- `--schema` - Database schema name (default: `public`)
- `--table` - Table name (required for entity/resource/module/all)
- `--tables` - Comma-separated table names (for builder command)
- `--version` - API version (default: `v1`)
- `--parts` - Module parts to generate (default: `handler,service,repository`)
- `--template-dir` - Custom template directory (overrides embedded templates)
- `--migrations` - Path to database migrations (default: `./db/migrations`)

### Builder-specific Flags
- `--new-module` - Generate complete new module
- `--dry-run` - Show what will be generated without writing files

## Output Structure

```
├── modules/
│   ├── {schema}/
│   │   ├── entity/
│   │   │   └── {table}.entity.go
│   │   ├── resource/
│   │   │   └── {table}.resource.go
│   │   └── {version}/
│   │       ├── handler/
│   │       ├── service/
│   │       └── repository/
│   └── {module}/
│       └── builder.go
├── app/
│   └── {module}_routes.go
├── common/
│   ├── cache/
│   │   └── redis.go (auto-generated cache keys)
│   └── constant/
│       └── permission.go (auto-generated permissions)
```

## Auto-Generated Features

### Cache Keys
When generating repositories, cache keys are automatically added to `common/cache/redis.go`:
```go
// UserFindByID is a redis key for find user by id.
UserFindByID = prefix + ":auth:user:find-by-id:%v"
// UserFindByName is a redis key for find user by name.
UserFindByName = prefix + ":auth:user:find-by-name:%v"
```

### Permission Constants
When generating routes, permission constants are automatically added to `common/constant/permission.go`:
```go
// User permissions
const (
    // PermUserView allows viewing user
    PermUserView = "user:view"
    // PermUserCreate allows creating user
    PermUserCreate = "user:create"
    // PermUserUpdate allows updating user
    PermUserUpdate = "user:update"
    // PermUserDelete allows deleting user
    PermUserDelete = "user:delete"
)
```

## Template System

The tool comes with built-in templates but supports full customization:

1. Run `starter-cli init` to copy embedded templates to `./templates`
2. Customize the templates in `./templates/`
3. Use `--template-dir=./templates` to use your custom templates

### Template Directory Structure
```
templates/
├── entity/
│   └── entity.tmpl
├── resource/
│   ├── resource.tmpl
│   ├── create_request.tmpl
│   └── update_request.tmpl
├── module/
│   ├── handler/
│   │   ├── creator.tmpl
│   │   ├── finder.tmpl
│   │   ├── updater.tmpl
│   │   └── deleter.tmpl
│   ├── service/
│   └── repository/
├── builder/
│   └── builder.tmpl
└── routes/
    └── routes.tmpl
```

## Requirements

- Go 1.21+
- SQL migration files in `./db/migrations/` (configurable via `--migrations`)
- Migration files should follow pattern: `{schema}/{table}.up.sql`

## Migration File Example

```sql
-- db/migrations/auth/users.up.sql
CREATE TABLE auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Support

If you encounter any issues or have questions:

1. Check the [issues page](https://github.com/rifqiakrm/starter-cli/issues)
2. Create a new issue with detailed description

---

**Happy coding!** 🎉