# Content Pack — Sample Training Documents

9 fictional corporate training documents across 3 topics, for use with the Question Generation Service take-home assignment.

## Structure

```
content-pack/
├── safety/
│   ├── safety-fire-emergency.md         — fire safety & evacuation procedures
│   ├── safety-hazard-reporting.md       — workplace hazard identification & reporting
│   └── safety-ppe-standards.md          — personal protective equipment standards
├── service/
│   ├── service-de-escalation.md         — de-escalation techniques for customer conflicts
│   ├── service-written-communication.md — effective written communication with customers
│   └── service-difficult-customers.md   — handling difficult customer types
└── privacy/
    ├── privacy-data-classification.md   — data classification & handling standards
    ├── privacy-phishing-defense.md      — phishing & social engineering defense
    └── privacy-access-control.md        — access control & password security
```

Each directory is a **topic**. Each `.md` file is a **content document** within that topic. The Question Generation Service should load these documents and generate quiz questions scoped to a given topic.

All content is fictional and created for this assignment — no real client data.
