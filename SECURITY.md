# Security Policy

Security is an important part of Vidya.

Vidya is designed to handle sensitive information such as student records, authentication credentials, school data, and payment-related information. Please report security vulnerabilities responsibly.

---

## Supported Versions

Vidya is currently under active development.

| Version              | Supported   |
| -------------------- | ----------- |
| `main`               | Yes         |
| Development releases | Best effort |
| Older releases       | No          |

Because the project is currently pre-`1.0.0`, security support may change as the project evolves.

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability, report it privately to the project maintainer.

Include as much of the following information as possible:

* Description of the vulnerability
* Affected component or module
* Steps required to reproduce the issue
* Potential security impact
* Proof of concept, if available
* Suggested mitigation, if known
* Relevant logs or screenshots

Please remove passwords, API keys, tokens, personal information, and other sensitive data from your report.

### Contact

**Maintainer:** Aditya Verma

Use the private contact method listed on the project's GitHub profile or repository documentation.

---

## What Should Be Reported Privately?

Examples include:

* Authentication bypasses
* Authorization vulnerabilities
* Cross-tenant data access
* SQL injection
* Remote code execution
* Cross-site scripting
* CSRF vulnerabilities
* Sensitive data exposure
* Credential or token leakage
* Payment security issues
* Razorpay webhook vulnerabilities
* Insecure file uploads
* Privilege escalation
* Server-side vulnerabilities
* Dependency vulnerabilities with meaningful impact
* Infrastructure vulnerabilities

When in doubt, report the issue privately.

---

## Responsible Disclosure

Please allow reasonable time for the vulnerability to be investigated and addressed before publicly disclosing technical details.

Do not intentionally exploit a vulnerability beyond what is reasonably necessary to demonstrate the issue.

Do not access, modify, delete, or disclose data belonging to other users or tenants.

---

## Security Updates

When appropriate, security fixes may be documented through:

* GitHub releases
* `CHANGELOG.md`
* Security advisories
* Relevant documentation

Sensitive exploit details will not be published if doing so could unnecessarily increase risk.

---

## Security Best Practices for Contributors

Contributors should:

* Never commit secrets or credentials.
* Never commit production database credentials.
* Never commit API keys or payment secrets.
* Never commit JWT signing keys.
* Use environment variables for secrets.
* Keep dependencies updated.
* Validate user-controlled input.
* Respect tenant boundaries.
* Follow the project's authentication and authorization architecture.
* Avoid logging sensitive information.
* Use secure defaults.

Never commit:

```text
.env
.env.production
private keys
API keys
JWT secrets
Razorpay production credentials
database passwords
access tokens
```

Use `.env.example` for documenting required configuration variables.

---

## Dependency Security

Dependencies should be reviewed before being introduced.

Contributors should consider:

* Maintenance status
* Known vulnerabilities
* License compatibility
* Dependency size
* Transitive dependencies
* Security implications

Security-related dependency updates should be prioritized when appropriate.

---

## Scope

This policy applies to:

* Vidya source code
* Backend services
* Flutter applications
* Database schema and migrations
* Deployment configuration
* CI/CD workflows
* Official project infrastructure
* Authentication and authorization systems
* Payment integrations

Third-party services and dependencies may have their own security reporting procedures.

---

## Thank You

Responsible security research helps make Vidya safer for schools, administrators, teachers, students, parents, and contributors.

Thank you for reporting vulnerabilities responsibly.
