# Project Governance

This document defines the governance model for the SaaSKit project. It establishes the roles, responsibilities, decision-making processes, and governance guidelines that keep SaaSKit healthy, active, and community-first.

---

## 1. Core Principles

SaaSKit is a community-driven, open-source project. We follow these core principles:
* **Openness:** All decisions, designs, and roadmaps are discussed publicly on GitHub.
* **Neutrality:** No single corporate entity controls the project. Decisions are based on technical merit and the best interest of the community.
* **Quality & Security:** We hold security and codebase stability as our highest priorities.
* **Developer Experience:** We maintain clear API boundaries and maintainable, readable code.

---

## 2. Roles and Responsibilities

We define three main levels of involvement in SaaSKit:

### A. Contributor
A Contributor is anyone who helps the project by submitting issues, participating in discussions, writing documentation, or submitting pull requests.
* **Prerequisites:** Agree to the [Code of Conduct](CODE_OF_CONDUCT.md) and sign off on all commits using the [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco).

### B. Maintainer
Maintainers have write access to the SaaSKit repositories and can merge pull requests, triage issues, and shape design decisions. They are stewards of the project.
* **Responsibilities:**
  - Actively review pull requests and issues.
  - Guide community contributions.
  - Maintain build quality, security standards, and documentation.
  - Participate in governance decisions.

### C. Steering Committee
The Steering Committee consists of senior maintainers responsible for:
  - Defining the overall project vision, roadmap, and release schedule.
  - Resolving technical disputes when maintainer consensus cannot be reached.
  - Administering project assets (domain names, keys, build infrastructure).

---

## 3. Maintainer Lifecycle

### Becoming a Maintainer
Contributors who demonstrate long-term commitment and high-quality contributions can be nominated as maintainers.
1. **Nomination:** Any existing maintainer can nominate a contributor.
2. **Review:** Maintainers review the nominee's contributions (quality of PRs, review behavior, community helpfulness).
3. **Approval:** A nomination is approved if two-thirds of active maintainers vote in favor, with no vetos, over a 7-day voting window.

### Inactivity and Retirement
To ensure the project moves forward, maintainers who are inactive for more than 6 months will be transitioned to "Emeritus Maintainer" status. Emeritus maintainers lose write/merge access but are permanently credited for their historical support. They can be reinstated by a simple majority vote of active maintainers.

---

## 4. Decision Making Process

SaaSKit operates on a **consensus-seeking** model. We prefer discussions that lead to agreement over formal voting.

### Steps to Decision
1. **Discussion:** Issues and ideas are raised in GitHub Issues, Discussions, or RFCs.
2. **Consensus:** If a proposal receives support and no objections from maintainers, it is accepted.
3. **Voting Fallback:** If consensus cannot be reached:
   - For standard code/PR decisions, a simple majority of maintainers is required.
   - For architectural changes (RFCs), a two-thirds majority is required.
   - For governance/policy changes, a two-thirds majority is required.

---

## 5. Request for Comments (RFC) Process

For major architectural, API, or behavioral changes, developers must follow the RFC process:
1. **Draft:** Create a markdown file detailing the proposal (Goals, Architecture, Implementation Details, Alternatives Considered, Trade-offs) in the `docs/rfcs/` folder or submit it as a PR.
2. **Review:** The community and maintainers review the draft. Discussions happen on the PR.
3. **Decision:** Once resolved, maintainers approve and merge the RFC. The implementation can then proceed, referencing the RFC number.

---

## 6. Developer Certificate of Origin (DCO)

To ensure clear chain of custody for open-source contributions, SaaSKit requires all commits to be signed off. This certifies that you have the right to submit the code under the Apache 2.0 license.

To sign off, add the `-s` or `--signoff` flag to your git commits:
```bash
git commit -s -m "auth: implement refresh token rotation"
```
This adds the following line to the end of your commit message:
```text
Signed-off-by: Jane Doe <jane.doe@example.com>
```
PRs containing unsigned commits will be automatically blocked by our CI validation.
