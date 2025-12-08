![SPIKE](../assets/spike-banner-lg.png)

```mermaid
graph TD
    A[Client requests secret] --> B[Extract peer SPIFFE ID]
    B --> C[Load all policies from cache]
    C --> D{For each policy}
    D --> E{SPIFFE ID matches pattern?}
    E -->|No| D
    E -->|Yes| F{Path matches pattern?}
    F -->|No| D
    F -->|Yes| G{Has required permission?}
    G -->|No| D
    G -->|Yes| H[Grant access]
    D -->|No matches| I[Deny access]
```
