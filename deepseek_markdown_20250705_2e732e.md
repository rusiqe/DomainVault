# DomainVault System Architecture

## Core Workflow
```mermaid
graph TD
    A[Registrar APIs] -->|API Polling| B(Provider Adapters)
    B --> C[Normalization Engine]
    C --> D[(Database)]
    D --> E[Sync Service]
    E --> F[Dashboard API]
    F --> G[Web UI]
```