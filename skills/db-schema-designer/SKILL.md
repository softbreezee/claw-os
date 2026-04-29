---
name: db-schema-designer
description: Design, create, inspect, and manage PostgreSQL tables for research and analysis data. Use when the user wants to store structured research results, asks to create a table, wants to see what tables exist, needs to understand the database schema, or wants to query/update research data.
metadata:
  fastclaw:
    always: false
    requires:
      env: []
---

# Database Schema Designer & Manager

You are the database authority. All PostgreSQL table operations go through this skill.

## Available Tools

| Tool | Purpose |
|------|---------|
| `db_create_table` | Create new tables (records to schema_registry) |
| `db_query` | SELECT / INSERT / UPDATE / DELETE (no DDL) |

DDL (CREATE/DROP/ALTER) is **only** allowed via `db_create_table`. `db_query` will reject DDL.

---

## Part 1: Schema Design — Before Creating Any Table

### Step 1: Identify Data Shape

Classify the data before writing any SQL:

| Shape | Characteristics | Recommendation |
|-------|----------------|----------------|
| **Entity** | Fixed attributes, one row = one object (company, product) | Standard relational table |
| **Time-series** | Metrics that change over time (revenue, stock price) | Add `period TEXT` ('2024Q3'), wide or EAV |
| **Hierarchical** | Tree structure (industry taxonomy, org chart) | `parent_id UUID` self-reference or `ltree` |
| **Relational** | Many-to-many (company↔product, company↔investor) | Junction table |
| **Unstructured** | Full-text, news, reports | `TEXT` + `tsvector` index |
| **Mixed** | Mostly fixed + some dynamic fields | Fixed columns + `extra JSONB` |

### Step 2: Apply Mandatory Standard Fields

Every table MUST have these fields:

```sql
id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
agent_id    TEXT        NOT NULL,                    -- which agent created this row
created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
source_url  TEXT,                                    -- data source (nullable)
embedding   vector(1536),                            -- for semantic search (nullable)
extra       JSONB       NOT NULL DEFAULT '{}'         -- flexible overflow attributes
```

### Step 3: Naming Convention

- **Table**: `research_{topic}_{type}` — e.g. `research_nvidia_companies`, `research_ev_metrics`
- **Fields**: `snake_case` English — e.g. `revenue_usd`, `employee_count`, `founded_year`
- **Avoid reserved words**: `value` → `metric_value`, `name` → `company_name`, `data` → `raw_data`

### Step 4: Index Strategy

```sql
-- Always
CREATE INDEX ON research_xxx (agent_id, created_at DESC);

-- For frequent filter fields
CREATE INDEX ON research_xxx (industry, founded_year);

-- Full-text search (when table has long text content)
CREATE INDEX ON research_xxx USING gin(to_tsvector('english', content));

-- Semantic search (only when embedding will be used AND rows > 1000)
CREATE INDEX ON research_xxx USING hnsw (embedding vector_cosine_ops);
```

### Step 5: Extensibility Rules

1. **Don't over-normalize**: Research data is read-heavy. Avoid deep JOINs.
2. **Use `extra JSONB`** for sparse or temporary attributes — don't add a column for every field.
3. **Time-series**: Use `period TEXT` ('2024Q3', '2024-03') not complex date types.
4. **Versioning**: If data needs history, add `version INT DEFAULT 1` and `superseded_by UUID`.

---

## Part 2: Create a Table

Always follow this sequence:

### 1. Explain the design to the user first
```
I'll create table `research_nvidia_companies` to store:
- Basic info: name, ticker, headquarters, founded_year
- Financials: revenue_usd, market_cap_usd  
- Standard fields: id, agent_id, embedding, extra, created_at, source_url
```

### 2. Call `db_create_table`

```sql
CREATE TABLE IF NOT EXISTS research_nvidia_companies (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id        TEXT        NOT NULL,
    company_name    TEXT        NOT NULL,
    ticker          TEXT,
    headquarters    TEXT,
    founded_year    INT,
    industry        TEXT,
    revenue_usd     NUMERIC,
    market_cap_usd  NUMERIC,
    employee_count  INT,
    description     TEXT,
    embedding       vector(1536),
    extra           JSONB       NOT NULL DEFAULT '{}',
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON research_nvidia_companies (agent_id, created_at DESC);
CREATE INDEX ON research_nvidia_companies (industry);
```

### 3. Verify after creation

```sql
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'research_nvidia_companies'
ORDER BY ordinal_position;
```

---

## Part 3: Inspect the Database

### See all agent-created tables
```sql
SELECT table_name, agent_id, purpose, created_at
FROM schema_registry
ORDER BY created_at DESC;
```

### See all tables including system tables
```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

### See columns of a specific table
```sql
SELECT column_name, data_type, character_maximum_length, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'research_nvidia_companies'
ORDER BY ordinal_position;
```

### See indexes on a table
```sql
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'research_nvidia_companies';
```

### See row counts across all research tables
```sql
SELECT schemaname, tablename,
       n_live_tup AS row_count,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;
```

---

## Part 4: Query & Manage Data

### Insert research data
```sql
INSERT INTO research_nvidia_companies
    (agent_id, company_name, ticker, industry, revenue_usd, source_url)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING id;
```
Pass values via `params` array, never inline them in SQL.

### Update existing rows
```sql
UPDATE research_nvidia_companies
SET revenue_usd = $1, extra = extra || $2::jsonb
WHERE id = $3;
```

### Semantic search (when embeddings are populated)
```sql
SELECT company_name, description, embedding <=> $1 AS distance
FROM research_nvidia_companies
WHERE agent_id = $2
ORDER BY embedding <=> $1
LIMIT 10;
```

### Full-text search
```sql
SELECT company_name, description
FROM research_nvidia_companies
WHERE to_tsvector('english', description) @@ plainto_tsquery('english', $1)
ORDER BY created_at DESC;
```

### Delete stale data
```sql
DELETE FROM research_nvidia_companies
WHERE created_at < now() - INTERVAL '90 days'
  AND agent_id = $1;
```

---

## Part 5: Maintenance

### Check for missing indexes (tables > 1000 rows without index)
```sql
SELECT relname AS table, seq_scan, idx_scan,
       pg_size_pretty(pg_total_relation_size(relid)) AS size
FROM pg_stat_user_tables
WHERE seq_scan > idx_scan AND n_live_tup > 1000
ORDER BY seq_scan DESC;
```

### Add a missing index on demand
Use `db_create_table` with just the CREATE INDEX statement:
```sql
CREATE INDEX IF NOT EXISTS research_xxx_field_idx ON research_xxx (field_name);
```
*(db_create_table allows CREATE INDEX in addition to CREATE TABLE)*

### Vacuum / analyze (after bulk inserts)
```sql
ANALYZE research_nvidia_companies;
```

---

## Decision Tree

```
User wants to store data?
  → Does a suitable table already exist?
      YES → INSERT with db_query
      NO  → Design schema → db_create_table → INSERT

User wants to find data?
  → Keyword search → db_query with ILIKE or tsvector
  → Semantic search → db_query with embedding <=>

User asks "what's in the database?"
  → db_query schema_registry + information_schema

User wants to clean up?
  → db_query DELETE with date filter
```
