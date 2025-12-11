---
description: Created with Workflow Studio
allowed-tools: Task,AskUserQuestion
---
```mermaid
flowchart TD
    start_node_default([Start])
    doc_check_agent[doc-check-agent]
    spec_review_agent[spec-review-agent]
    implementation_agent[implementation-agent]
    build_agent[build-agent]
    test_agent[test-agent]
    doc_update_agent[doc-update-agent]
    end_node_default([End])

    start_node_default --> doc_check_agent
    doc_check_agent --> spec_review_agent
    spec_review_agent --> implementation_agent
    implementation_agent --> build_agent
    build_agent --> test_agent
    test_agent --> doc_update_agent
    doc_update_agent --> end_node_default
```

## Workflow Execution Guide

Follow the Mermaid flowchart above to execute the workflow. Each node type has specific execution methods as described below.

### Execution Methods by Node Type

- **Rectangle nodes**: Execute Sub-Agents using the Task tool
- **Diamond nodes (AskUserQuestion:...)**: Use the AskUserQuestion tool to prompt the user and branch based on their response
- **Diamond nodes (Branch/Switch:...)**: Automatically branch based on the results of previous processing (see details section)
- **Rectangle nodes (Prompt nodes)**: Execute the prompts described in the details section below
