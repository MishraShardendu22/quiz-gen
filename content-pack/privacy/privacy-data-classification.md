# Data Classification and Handling Standards

## Overview

Every piece of information in an organization carries a different level of sensitivity. Treating all data the same way is both inefficient (over-protecting low-sensitivity data wastes resources) and dangerous (under-protecting high-sensitivity data creates breach risk). This module defines the four data classification levels and the handling rules that apply to each.

## Classification Levels

### Level 1: Public

Information that is already approved for public consumption or would cause no harm if disclosed.
- Published marketing materials, press releases
- Job postings
- Product documentation available to customers
- Publicly announced business partnerships

**Handling:** No restrictions on storage, transmission, or disposal. May be shared externally without approval. No encryption required (though it is permitted).

### Level 2: Internal

Information meant for internal use that is not sensitive but could cause minor embarrassment or operational inconvenience if disclosed externally.
- Internal memos and meeting notes
- Non-sensitive operational procedures
- General employee directory information (names, roles, department)
- Internal training materials

**Handling:** Store on company-managed systems. Do not post on external platforms. No encryption required for storage, but use company email (not personal) for transmission. Shred paper copies before disposal.

### Level 3: Confidential

Information that could cause significant harm to the company, employees, or customers if disclosed. This is the most common classification for business data.
- Customer contact information and order history
- Employee personal information (addresses, phone numbers, compensation)
- Financial data (revenue figures, pricing strategies, cost structures)
- Vendor contracts and supplier pricing
- Internal security procedures and network architecture

**Handling:** Encrypt at rest and in transit. Access restricted to employees with a legitimate business need — access must be approved by a manager. Do not store on local devices (laptops, phones) unless full-disk encryption is enabled. Do not email externally without management approval. When sharing internally, use the principle of least privilege — share with the specific individuals who need it, not with entire departments. Securely shred all paper copies.

### Level 4: Restricted

The highest sensitivity level. Disclosure could cause severe financial, legal, or reputational damage, or violate regulatory obligations.
- Payment card data (card numbers, CVVs) — subject to PCI-DSS
- Government identification numbers (Aadhaar, SSN, passport numbers)
- Health information — subject to HIPAA or equivalent local regulations
- Authentication credentials (passwords, private keys, API secrets)
- Legal documents under attorney-client privilege
- M&A and pre-earnings-release financial data

**Handling:** Encrypt at rest with individual-key encryption (not just volume-level). Access restricted to named individuals, reviewed quarterly. No transmission by email under any circumstances — use the secure file transfer platform only. No local storage on any device. No printing without explicit, documented, time-limited approval from the data owner. Access is logged and audited monthly. Violations are subject to immediate disciplinary action and potential legal consequences.

## Classification Responsibilities

### Data Owners
Each data category has a designated owner (typically a department head) responsible for:
- Assigning the correct classification level
- Approving access requests
- Reviewing access lists quarterly
- Ensuring handling rules are communicated to all users of the data

### Data Custodians
IT and Security teams are custodians responsible for:
- Implementing the technical controls required by each classification level
- Monitoring access logs for Level 3 and 4 data
- Conducting periodic audits of storage and transmission practices
- Reporting violations to the data owner and Security Committee

### All Employees
Every employee who handles data is responsible for:
- Knowing the classification level of the data they work with
- Following the handling rules for that level
- Reporting suspected misclassification to the data owner
- Reporting any data handling violation to Security within 1 hour of discovery

## Common Classification Mistakes

1. **Defaulting to Confidential for everything** — this defeats the purpose of classification and creates alert fatigue. If everything is confidential, nothing gets the extra protection that truly sensitive data requires.
2. **Classifying by format instead of content** — a spreadsheet is not automatically "Internal" and a database is not automatically "Confidential." The content determines the level, not the container.
3. **Inheriting classification from source** — a public press release quoted in an internal strategy memo does not make the memo "Public." The memo is classified based on its overall content and context.
4. **Failing to reclassify** — data that was Internal during project planning may become Confidential once the project launches. Review classifications when context changes.
