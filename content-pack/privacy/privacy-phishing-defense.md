# Phishing and Social Engineering Defense

## What is Social Engineering?

Social engineering is the manipulation of people into performing actions or divulging confidential information. Unlike technical attacks that exploit software vulnerabilities, social engineering exploits human psychology — trust, authority, urgency, helpfulness, and fear. It is the most common attack vector in corporate breaches, because it is far easier to trick a person than to break a firewall.

## Types of Social Engineering Attacks

### Phishing
Fraudulent emails that appear to come from a trusted source. The email typically contains either a malicious link (directing to a fake login page that captures credentials) or a malicious attachment (installing malware when opened).

**Red flags in phishing emails:**
- Urgent or threatening language: "Your account will be suspended in 24 hours"
- Generic greetings: "Dear Customer" instead of your name
- Mismatched sender domains: email claims to be from IT but the domain is `it-support-external.com`
- Unexpected attachments, especially .zip, .exe, or .js files
- Links that do not match the displayed text — hover over links to see the actual URL before clicking
- Requests for credentials, payment card information, or sensitive data via email — no legitimate internal team will ask for these by email

### Spear Phishing
Targeted phishing aimed at a specific individual, using personal details to appear credible. The attacker may know your name, role, projects you're working on, and your manager's name — information gathered from LinkedIn, company websites, or prior breaches. Spear phishing is harder to detect because the personalization makes it convincing.

### Vishing (Voice Phishing)
Phone calls from attackers pretending to be IT support, vendors, executives, or government officials. Common scenarios:
- "This is IT support — we've detected a virus on your machine. I need your password to run a cleanup tool remotely."
- "This is the IRS/tax authority — you owe back taxes and a warrant has been issued. Pay now or face arrest."
- "This is [executive name]'s assistant — [executive] is in a meeting and needs you to urgently process a wire transfer."

### Smishing (SMS Phishing)
Text messages with urgent requests or malicious links. Example: "Your package delivery failed — click here to reschedule" with a link to a credential-harvesting page.

### Tailgating and Physical Social Engineering
An attacker follows an authorized employee through a secured door, often carrying heavy boxes and relying on the employee's courtesy to hold the door. Other physical techniques: impersonating a vendor or auditor, leaving infected USB drives in the parking lot or lobby (curiosity leads someone to plug them in), or posing as a fire inspector to access server rooms.

## The Five Psychological Triggers

Social engineers exploit five universal psychological tendencies. Recognizing them is your primary defense:

1. **Authority** — we comply with people we perceive as having authority. An email "from the CEO" or a call "from IT" triggers compliance. **Defense:** Verify identity through a secondary channel. If the CEO emails you asking for an urgent wire transfer, call the CEO's office to confirm. No legitimate authority figure will object to verification.

2. **Urgency** — time pressure prevents careful thinking. "You must act in the next 10 minutes" or "This offer expires today." **Defense:** Urgency is the single most reliable phishing signal. Any communication demanding immediate action should trigger suspicion, not compliance. Legitimate business rarely requires action in minutes.

3. **Familiarity** — we trust people we recognize. An email that references a colleague's name or a project you're working on feels safer. **Defense:** Familiarity is not authentication. The fact that an email mentions your project does not mean it came from your project team. Check the sender's actual email address, not just the display name.

4. **Helpfulness** — our instinct to be helpful is exploited by attackers posing as someone in need. "I'm locked out of my account and I have a presentation in 5 minutes — can you reset my password?" **Defense:** Follow procedures even when helping. If the password reset process requires identity verification, perform it regardless of the person's urgency. A real colleague will understand; an attacker will try to bypass.

5. **Fear** — threats of consequences trigger compliance. "Your account will be terminated" or "Legal action will be taken." **Defense:** Fear-based communications should always be verified. No legitimate organization terminates accounts or initiates legal action without prior written notice through official channels.

## Reporting and Response

If you suspect you have received a social engineering attempt:

1. **Do not click links, open attachments, or provide information.**
2. **Do not delete the message** — preserve it for the Security team's investigation.
3. **Report it:** forward suspicious emails to `security@company.internal`, report suspicious calls or texts to Security at ext. 7700.
4. **If you already clicked or provided information:** report immediately to Security — speed matters. A credential stolen 5 minutes ago can be revoked; one stolen 5 days ago may already be in use. You will not be penalized for reporting a mistake — you will be penalized for hiding one.
