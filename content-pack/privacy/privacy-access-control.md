# Access Control and Password Security

## Principle of Least Privilege

Every employee should have the minimum level of access required to perform their job — no more. This principle, called Least Privilege, is the foundation of access control. If an employee does not need access to a system, file, or database for their daily work, they should not have it.

**Why least privilege matters:**
- If your account is compromised, the attacker can only access what you can access. An employee with access to 3 systems gives an attacker 3 systems. An employee with access to 30 gives an attacker 30.
- It limits accidental damage — you cannot delete data you cannot access.
- It simplifies audits — fewer access points to review and justify.

**Your responsibilities:**
- Request access only when you have a specific, current business need — not "in case I need it someday"
- Request access removal when you change roles or no longer need a system
- Review your access list quarterly when prompted by IT and justify each system you retain

## Password Standards

### Length Over Complexity
Current guidance from NIST and other standards bodies has shifted: password length is more important than complexity. A 16-character passphrase like `correct-horse-battery-staple` is stronger than an 8-character password like `Tr@v3l!1` and is easier to remember.

**Minimum requirements:**
- Minimum 14 characters for all systems
- Must not be a common password or a known breached password (the system checks against breach databases automatically)
- No maximum length — use passphrases as long as you like
- Special characters are permitted but not required — length is the priority

### Password Management
- **Never reuse passwords across systems.** If one system is breached, attackers will try the same password on other systems (credential stuffing).
- **Use a company-approved password manager.** Password managers generate and store unique, strong passwords for every site. You only need to remember one master passphrase. Never store passwords in browsers (they are vulnerable to malware extraction), in spreadsheets, or in notes files.
- **Never share passwords with anyone, including IT.** No legitimate IT staff member will ever ask for your password. If someone asks, it is a social engineering attempt — report it.
- **Change passwords immediately if you suspect compromise.** Do not wait for the next scheduled rotation.

### Multi-Factor Authentication (MFA)

MFA is mandatory for all company accounts. It adds a second factor beyond your password — something you have (a mobile authenticator app, a hardware key) or something you are (biometric).

**MFA methods, strongest to weakest:**
1. **Hardware security keys (FIDO2 / WebAuthn)** — the gold standard. Phishing-resistant because the key cryptographically verifies the website domain. Use for high-privilege accounts.
2. **Authenticator apps (TOTP)** — e.g. Google Authenticator, Authy, Microsoft Authenticator. Strong, but the code can be phished if you are tricked into entering it on a fake site.
3. **Push notifications** — a prompt appears on your phone asking you to approve or deny the login. Convenient, but vulnerable to "MFA fatigue" attacks (the attacker repeatedly sends login attempts hoping you'll approve one out of annoyance). If you receive an MFA prompt you did not initiate, deny it and report it to Security.
4. **SMS codes** — the weakest MFA method. SIM swapping attacks can redirect your SMS messages to an attacker. SMS is better than no MFA, but upgrade to an authenticator app or hardware key as soon as possible.

### MFA Fatigue Attacks

An emerging attack pattern: the attacker has your password (from a breach) and sends repeated MFA push notifications to your phone, sometimes at 2 AM or during meetings, hoping you'll approve one to stop the notifications. If you receive an unexpected MFA prompt:
1. **Deny it.** Always.
2. **Do not approve it "just to make it stop."** One approval gives the attacker full access.
3. **Report it to Security** — repeated unsolicited MFA prompts indicate your password may be compromised and needs changing.

## Session Management

- **Lock your screen when you step away** — even for a minute. Win+L (Windows) or Ctrl+Cmd+Q (Mac). Unlocked screens are the easiest way for someone to access your account in a physical office.
- **Log out of shared or public computers entirely** — do not just close the browser. Clear the browser data if possible.
- **Do not save login sessions on devices you do not control** — uncheck "Remember me" on any device that is not yours.
- **Review active sessions periodically** — most platforms show where you are logged in. If you see a session from a location or device you do not recognize, terminate it and change your password.

## Access Reviews

IT conducts quarterly access reviews. You will receive a list of all systems and data you have access to and must confirm that each one is still required for your current role. This is not a formality — stale access is one of the most common audit findings and a significant security risk. If your role has changed and you still have access to systems from your previous role, flag it for removal during the review.
