# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository is a **Claude Skills development framework** for creating, testing, and packaging custom skills that Claude can use to automate specific tasks. Skills bundle instructions, reference materials, code templates, and examples into a structured format.

## Core Architecture

### Skill Structure
Every skill follows this standardized directory layout:

```
skill-name/
├── SKILL.md          # Required: Metadata (name, description) + instructions
├── references/       # Optional: Design guidelines, patterns, documentation
├── scripts/          # Optional: Code templates, utilities
└── examples/         # Optional: Usage examples, sample prompts
```

### SKILL.md Format
The entry point for every skill with YAML frontmatter:

```markdown
---
name: skill-name
description: Brief description of when and how to use this skill
---

# Skill Name

[Instructions for Claude to follow]

## Workflow
- Step-by-step process
```

**Critical**: The `name` field in frontmatter determines how users invoke the skill.

## Development Workflow

### Creating a New Skill

1. **Structure Setup**: Create directory with `SKILL.md` as minimum requirement
2. **Write Instructions**: Define clear, actionable steps Claude should follow
3. **Add Supporting Materials**:
   - `references/`: Design patterns, API specs, guidelines
   - `scripts/`: Reusable code templates with inline documentation
   - `examples/`: Sample prompts showing skill invocation
4. **Package**: ZIP the skill folder for distribution

### Skill Design Principles

- **Specificity**: Skills should target concrete use cases (API development, error handling, testing)
- **Completeness**: Include all context needed - don't assume external knowledge
- **Template-Driven**: Provide code templates in `scripts/` that demonstrate patterns
- **Reference Integration**: Link to relevant files using relative paths

### Example Skill Analysis

The included `example-skill/` demonstrates backend API helper patterns:

- **SKILL.md**: Defines RESTful principles, error handling workflow, response formats
- **references/api-design-guidelines.md**: Resource naming, HTTP methods, status codes, versioning strategy
- **scripts/api-template.js**: Express.js template with validation, business logic placeholders, consistent error handling
- **examples/usage-example.md**: Sample prompts for endpoint creation, error handling improvements, design reviews

## Skill Invocation Pattern

Users invoke skills via natural language:
```
"Use the 'backend-api-helper' skill to create a GET endpoint for user profiles"
```

Claude will automatically load `SKILL.md` and reference supporting materials from the skill folder.

## Key Design Patterns

### API Response Standardization
All API-related skills follow consistent JSON response structure:

**Success**:
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

**Error**:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

### RESTful Conventions
- **URLs**: `/api/v{version}/{resources}` (plural nouns, lowercase, hyphens)
- **Methods**: GET (read), POST (create), PUT (full update), PATCH (partial), DELETE
- **Status Codes**: 200 (OK), 201 (Created), 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 404 (Not Found), 500 (Server Error)

### Security Baseline
All code templates incorporate:
- Request parameter validation before business logic
- Input sanitization to prevent injection attacks
- Proper authentication token verification
- Comprehensive error logging without exposing internals

## Development Guidelines

### When Creating Skills
- **Single Responsibility**: Each skill should address one domain or task type
- **Actionable Instructions**: Write step-by-step workflows, not abstract concepts
- **Code Quality**: Templates should be production-ready, not pseudocode
- **Localization**: Current examples use Korean for descriptions - maintain consistency

### File Organization
- Keep `references/` focused on design decisions and architectural patterns
- Use `scripts/` for concrete, copy-paste-ready code (not pseudo-code)
- Make `examples/` demonstrate actual user prompts, not theoretical scenarios

## Extension Points

To add a new skill domain:
1. Create `{skill-name}/SKILL.md` with clear problem scope
2. Document design decisions in `references/`
3. Provide working templates in `scripts/`
4. Show usage patterns in `examples/`

The framework is unopinionated about technology stack - adapt structure to any language/framework.
