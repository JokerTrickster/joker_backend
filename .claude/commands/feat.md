---
description: Created with Workflow Studio
allowed-tools: Task,AskUserQuestion
---
```mermaid
flowchart TD
    start_node_default([Start])
    doc_writer_1[doc-writer-1]
    spec_reviewer_1[spec-reviewer-1]
    backend_dev_1[backend-dev-1]
    test_agent_1[test-agent-1]
    test_check_1{If/Else:<br/>Conditional Branch}
    doc_updater_1[doc-updater-1]
    fix_agent_1[fix-agent-1]
    retest_1[retest-1]
    end_node_success([End])
    end_node_retry([End])

    start_node_default --> doc_writer_1
    doc_writer_1 --> spec_reviewer_1
    spec_reviewer_1 --> backend_dev_1
    backend_dev_1 --> test_agent_1
    test_agent_1 --> test_check_1
    test_check_1 -->|Pass| doc_updater_1
    test_check_1 -->|Fail| fix_agent_1
    doc_updater_1 --> end_node_success
    fix_agent_1 --> retest_1
    retest_1 --> end_node_retry
```

## Workflow Execution Guide

Follow the Mermaid flowchart above to execute the workflow. Each node type has specific execution methods as described below.

### Execution Methods by Node Type

- **Rectangle nodes**: Execute Sub-Agents using the Task tool
- **Diamond nodes (AskUserQuestion:...)**: Use the AskUserQuestion tool to prompt the user and branch based on their response
- **Diamond nodes (Branch/Switch:...)**: Automatically branch based on the results of previous processing (see details section)
- **Rectangle nodes (Prompt nodes)**: Execute the prompts described in the details section below

### If/Else Node Details

#### test_check_1(Binary Branch (True/False))

**Evaluation Target**: Test and build results

**Branch conditions:**
- **Pass**: All tests pass and build succeeds
- **Fail**: Tests fail or build errors exist

**Execution method**: Evaluate the results of the previous processing and automatically select the appropriate branch based on the conditions above.
