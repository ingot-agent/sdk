# Security policy: Ingot SDK

## Reporting a suspected vulnerability

Do not post exploit details, credentials, private prompts, state files or
sensitive reproduction data in a public issue or pull request.

**Reporting setup checked on 2026-09-22:** GitHub's API reported private
vulnerability reporting disabled for this repository. No dedicated security
email or guaranteed response schedule is published in this checkout.

Open the repository's [Security page](https://github.com/ingot-agent/sdk/security).
If maintainers have since enabled **Report a vulnerability**, use that private
form. Otherwise, open a [contact request](https://github.com/ingot-agent/sdk/issues/new?title=Private%20security%20reporting%20contact%20request)
containing only: “Please provide a private channel for a security report.”
Do not include affected code paths, reproduction steps or attachments in that
public coordination request. Wait for maintainers to establish a private
channel before sending technical details.

Once a private channel is available, include:

- exact SDK module version or source commit, affected package/API, Go version, OS/architecture, and versions of relevant consumers;
- the impact and the access/conditions needed to reproduce it;
- the smallest reproduction with synthetic data and credentials;
- expected versus observed behavior, and any suggested mitigation.

If the report spans repositories, name all affected modules in one private
report so maintainers can coordinate it. Ordinary non-security defects belong
in the public bug-report form.

## Scope and trust boundaries

The SDK defines public capability contracts, including content, tool/model
invocation, execution scope, interaction, Operations and observations. Report
contract defects that permit incorrect routing, unintended data exposure or
unsafe ownership/concurrency behavior here. Concrete host authorization and
provider/tool implementations belong to the owning plugin repository.

An execution scope expresses caller routing; it is not a credential, tenant
boundary or OS sandbox. Contract validation cannot replace authentication and
authorization in an application host. Sensitive field metadata does not
implement encryption or a secret store. Public contracts must not acquire
business routing authority from hidden context values.

## Version information and disclosure

Please report the exact affected versions, including older releases and source
checkouts. This repository does not currently publish an LTS or guaranteed
security-backport schedule. Fix availability and migration requirements must
be stated in the corresponding release notes; an unreleased branch fix should
not be described as available in an existing tag.

Use the established private channel to coordinate investigation and disclosure.
Do not assume this document guarantees a response deadline or authorizes
testing systems, accounts or data that you do not control.

## Maintainer release requirement

Before public release, enable **Private vulnerability reporting** in the GitHub
repository's security settings, verify the reporter-facing form with an
appropriate account, and update the dated setup statement above. Monitor the
chosen channel and document any support/response policy only after it has been
agreed. Adding this file or an issue-template link does not enable reporting
in GitHub settings.

