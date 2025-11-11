# Documentation Index

Complete documentation for the Weather Data Collector Service.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Quick Navigation

### Getting Started

- **[README](../README.md)** - Main documentation with overview, quick start, and features
- **[Quick Start Guide](../README.md#quick-start)** - Get up and running in 5 minutes
- **[Configuration Reference](CONFIGURATION.md)** - Complete configuration documentation

### Development

- **[Development Guide](DEVELOPMENT.md)** - Setup, testing, code organization, git workflow
- **[API Documentation](API.md)** - Internal service interfaces and data models
- **[Architecture Documentation](ARCHITECTURE.md)** - System design, components, data flow

### Operations

- **[Operations Runbook](RUNBOOK.md)** - Restart procedures, scaling, debugging, monitoring checklist
- **[Troubleshooting Guide](TROUBLESHOOTING.md)** - Symptom-based problem solving
- **[Deployment Guide](DEPLOYMENT.md)** - Docker, Kubernetes, systemd deployment
- **[Monitoring Guide](../MONITORING.md)** - Metrics, Grafana dashboards, alerts

### Reference

- **[Configuration Reference](CONFIGURATION.md)** - All environment variables documented
- **[Architecture Diagrams](diagrams/)** - Visual system architecture

## Documentation Structure

```
docs/
├── INDEX.md                          # This file
├── API.md                           # Internal API documentation
├── ARCHITECTURE.md                  # System architecture
├── RUNBOOK.md                       # Operations procedures
├── CONFIGURATION.md                 # Configuration reference
├── TROUBLESHOOTING.md               # Problem solving
├── DEVELOPMENT.md                   # Development guide
├── DEPLOYMENT.md                    # Deployment guides
└── diagrams/
    ├── system-overview.mmd          # High-level system diagram
    ├── component-architecture.mmd   # Component relationships
    ├── sequence-alarm-processing.mmd # Alarm processing flow
    └── deployment-architecture.mmd  # Kubernetes deployment
```

## Documentation by Role

### For Developers

1. [Development Guide](DEVELOPMENT.md) - Setup and workflow
2. [API Documentation](API.md) - Interface specifications
3. [Architecture Documentation](ARCHITECTURE.md) - System design
4. [Testing Guide](DEVELOPMENT.md#testing) - Unit and integration tests

### For DevOps/SRE

1. [Operations Runbook](RUNBOOK.md) - Day-to-day operations
2. [Deployment Guide](DEPLOYMENT.md) - Deployment procedures
3. [Monitoring Guide](../MONITORING.md) - Observability setup
4. [Troubleshooting Guide](TROUBLESHOOTING.md) - Problem resolution

### For Product/Business

1. [README](../README.md) - Product overview and features
2. [Architecture Overview](ARCHITECTURE.md#system-overview) - High-level design
3. [Performance Characteristics](ARCHITECTURE.md#performance-characteristics) - Capacity planning

## Documentation by Task

### I want to...

**...get started quickly**
→ [Quick Start Guide](../README.md#quick-start)

**...understand the architecture**
→ [Architecture Documentation](ARCHITECTURE.md)
→ [System Overview Diagram](diagrams/system-overview.mmd)

**...deploy to production**
→ [Deployment Guide](DEPLOYMENT.md)
→ [Configuration Reference](CONFIGURATION.md)

**...troubleshoot an issue**
→ [Troubleshooting Guide](TROUBLESHOOTING.md)
→ [Operations Runbook](RUNBOOK.md)

**...add a new feature**
→ [Development Guide](DEVELOPMENT.md)
→ [API Documentation](API.md)

**...configure the service**
→ [Configuration Reference](CONFIGURATION.md)
→ [.env.example](../.env.example)

**...monitor the service**
→ [Monitoring Guide](../MONITORING.md)
→ [Metrics Reference](RUNBOOK.md#monitoring-checklist)

**...scale the service**
→ [Scaling Strategies](RUNBOOK.md#scaling-strategies)
→ [Performance Tuning](RUNBOOK.md#performance-tuning)

**...restart the service**
→ [Restart Procedures](RUNBOOK.md#restart-procedures)

**...understand the interfaces**
→ [API Documentation](API.md)
→ [Component Architecture](diagrams/component-architecture.mmd)

## Documentation Standards

### Formatting

- Use Markdown for all documentation
- Use Mermaid for diagrams
- Follow semantic line breaks
- Include code examples with syntax highlighting
- Use tables for structured data

### Structure

Each documentation file should include:
- Title and version
- Table of contents (for files > 200 lines)
- Clear sections with headers
- Code examples where applicable
- Cross-references to related docs

### Maintenance

- Update documentation with code changes
- Keep version numbers in sync
- Review quarterly for accuracy
- Archive outdated versions

## External Resources

### Official Documentation

- **Go:** https://go.dev/doc/
- **GORM:** https://gorm.io/docs/
- **Redis:** https://redis.io/docs/
- **Firebase:** https://firebase.google.com/docs/cloud-messaging
- **Prometheus:** https://prometheus.io/docs/
- **Kubernetes:** https://kubernetes.io/docs/

### Related Projects

- **Joker Backend:** https://github.com/JokerTrickster/joker_backend
- **Firebase Go SDK:** https://firebase.google.com/docs/admin/setup

## Contributing to Documentation

### Documentation Changes

1. Update relevant documentation files
2. Update version and "Last Updated" date
3. Update this index if structure changes
4. Test code examples
5. Review for clarity and accuracy

### Submitting Changes

```bash
# Create feature branch
git checkout -b docs/update-configuration

# Make changes
vim docs/CONFIGURATION.md

# Commit with conventional commit
git commit -m "docs(config): add new environment variable documentation"

# Push and create PR
git push origin docs/update-configuration
```

## Version History

- **v1.0.0** (2025-11-11): Initial documentation release
  - README.md with overview and quick start
  - Complete API documentation
  - Architecture documentation with diagrams
  - Operations runbook
  - Configuration reference
  - Troubleshooting guide
  - Development guide
  - Deployment guide for Docker, Kubernetes, systemd
  - Mermaid diagrams for system, component, sequence, deployment

## Contact

- **GitHub Issues:** https://github.com/JokerTrickster/joker_backend/issues
- **Documentation Feedback:** Create issue with `documentation` label
- **Email:** support@example.com
