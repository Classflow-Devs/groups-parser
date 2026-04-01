# groups-parser

A CLI tool that scrapes academic groups from the [MGUTM](https://dec.mgutm.ru) student portal and persists them into a PostgreSQL database as part of the **ClassFlow** platform.

## How it works

1. Fetches the full list of groups for a given academic year from the portal API.
2. Concurrently fetches detailed info for each group (faculty, speciality, education level) using a configurable worker pool.
3. Auto-detects the target branch by matching the first two letters of the group name against `branch.city` in the database. Falls back to branch ID `1` when no match is found.
4. Upserts `Faculty` and `Speciality` records, then inserts new `Group` records — skipping any that already exist.

All HTTP requests use exponential back-off retry (up to 3 attempts, 2 s → 4 s delay) to handle transient timeouts from the portal.

## Database schema

```
branches
  ├── faculties    (branch_id → branches.id)
  ├── specialities (branch_id → branches.id)
  └── groups       (branch_id, faculty_id, speciality_id)
```

The `branches` table must be populated before running the parser. The parser creates faculties, specialities, and groups automatically.

### Branch auto-detection

The parser resolves the branch for each group by extracting the first two uppercase letters of the group name (e.g. `"АИС-301"` → `"АИ"`) and looking for a `Branch` record whose `country` field matches. If no match is found, the group is assigned to branch ID `1`.

To map a branch, set its `country` column to the two-letter prefix used by that branch's group names.

## Build

```bash
go build -o groups-parser .
```

Requires Go 1.21+.

## Usage

```
./groups-parser [flags]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--dsn` | yes | — | PostgreSQL connection string |
| `--year` | no | `2025-2026` | Academic year to parse |
| `--workers` | no | `20` | Number of concurrent HTTP workers |
| `--migrate` | no | `false` | Run `AutoMigrate` before parsing |

### Examples

First run (with migration):

```bash
./groups-parser \
  --dsn "postgres://postgres:secret@localhost:5432/classflow" \
  --year 2025-2026 \
  --migrate
```

Subsequent runs:

```bash
./groups-parser \
  --dsn "postgres://postgres:secret@localhost:5432/classflow"
```

Sample output:

```
2025/04/01 12:00:00 loaded 3 branches
2025/04/01 12:00:01 fetched 412 groups for year 2025-2026
2025/04/01 12:00:45 fetched detailed info for 412/412 groups
2025/04/01 12:00:45 done — saved: 398  skipped: 14  errors: 0
```

## License

[MIT](LICENSE)