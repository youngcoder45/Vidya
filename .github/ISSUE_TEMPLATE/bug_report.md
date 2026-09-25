## `.github/ISSUE_TEMPLATE/bug_report.md`

````markdown
---
name: Bug Report
about: Report a reproducible problem in SchoolOS
title: "[Bug]: "
labels: bug
assignees: ""
---

# Bug Report

## Description

<!-- Clearly describe the problem. -->

## Expected Behavior

<!-- What did you expect to happen? -->

## Actual Behavior

<!-- What actually happened? -->

## Steps to Reproduce

1.
2.
3.
4.

## Environment

### Backend

```text
Go version:
OS:
Database:
Redis:
SchoolOS commit/version:
````

### Flutter

```text
Flutter version:
Dart version:
OS:
Device:
SchoolOS commit/version:
```

## Logs / Error Messages

<!--
Paste relevant logs or error messages here.

DO NOT include:
- Passwords
- API keys
- JWTs
- Database credentials
- Personal information
- Production secrets
-->

```text
```

## Screenshots / Videos

<!-- Add screenshots or screen recordings if they help demonstrate the issue. -->

## Reproduction Repository

<!-- If applicable, provide a minimal reproduction repository or branch. -->

## Additional Context

<!-- Add any other relevant information. -->

## Security Disclosure

If this issue involves a security vulnerability, **do not submit it publicly**.

Please follow [`SECURITY.md`](../../SECURITY.md) instead.

````

---

## `.editorconfig`

```ini
# EditorConfig helps maintain consistent coding styles
# across different editors and IDEs.

root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
indent_style = space
indent_size = 2
trim_trailing_whitespace = true

[*.md]
trim_trailing_whitespace = false
max_line_length = off

[*.go]
indent_style = tab
indent_size = 4

[*.dart]
indent_style = space
indent_size = 2

[*.yml]
indent_style = space
indent_size = 2

[*.yaml]
indent_style = space
indent_size = 2

[*.json]
indent_style = space
indent_size = 2

[*.sql]
indent_style = space
indent_size = 2

[Makefile]
indent_style = tab
````
