# RBAC ER Diagram

Төмендегі диаграмма 14-суретке арналған: `users`, `roles`, `permissions`, `role_permissions`, `role_assignments` және оларға жалғанатын `tenant`, `faculty`, `department`, `project` контекстері.

```mermaid
erDiagram
  USERS {
    UUID id PK
    UUID tenant_id FK
    TEXT email
    TEXT status
    TIMESTAMPTZ created_at
  }

  TENANTS {
    UUID id PK
    TEXT code
    TEXT name
    TEXT status
  }

  FACULTIES {
    UUID id PK
    UUID tenant_id FK
    TEXT code
    TEXT name
  }

  DEPARTMENTS {
    UUID id PK
    UUID tenant_id FK
    UUID faculty_id FK
    TEXT code
    TEXT name
  }

  PROJECTS {
    UUID id PK
    UUID tenant_id FK
    UUID faculty_id FK
    TEXT title
    TEXT status
  }

  ROLES {
    UUID id PK
    TEXT code
    TEXT name
    TIMESTAMPTZ created_at
  }

  PERMISSIONS {
    UUID id PK
    TEXT code
    TEXT description
    TIMESTAMPTZ created_at
  }

  ROLE_PERMISSIONS {
    UUID role_id PK,FK
    UUID permission_id PK,FK
  }

  ROLE_ASSIGNMENTS {
    UUID id PK
    UUID tenant_id FK
    UUID user_id
    UUID role_id FK
    TEXT scope_type
    UUID scope_id
    TIMESTAMPTZ expires_at
    TIMESTAMPTZ created_at
  }

  TENANTS ||--o{ USERS : contains
  TENANTS ||--o{ FACULTIES : contains
  TENANTS ||--o{ DEPARTMENTS : contains
  TENANTS ||--o{ PROJECTS : contains
  TENANTS ||--o{ ROLE_ASSIGNMENTS : scopes

  FACULTIES ||--o{ DEPARTMENTS : contains
  FACULTIES ||--o{ PROJECTS : scopes

  USERS ||--o{ ROLE_ASSIGNMENTS : receives
  ROLES ||--o{ ROLE_ASSIGNMENTS : assigned_as

  ROLES ||--o{ ROLE_PERMISSIONS : has
  PERMISSIONS ||--o{ ROLE_PERMISSIONS : included_in

  FACULTIES ||..o{ ROLE_ASSIGNMENTS : "scope_id when FACULTY"
  DEPARTMENTS ||..o{ ROLE_ASSIGNMENTS : "scope_id when DEPARTMENT"
  PROJECTS ||..o{ ROLE_ASSIGNMENTS : "scope_id when PROJECT"
```

`role_permissions` кестесі `roles` пен `permissions` арасындағы many-to-many байланысты береді. `role_assignments` кестесі нақты пайдаланушыға белгілі бір рөлді тағайындайды және `scope_type`/`scope_id` арқылы оның контекстін анықтайды: tenant, faculty, department немесе project.
