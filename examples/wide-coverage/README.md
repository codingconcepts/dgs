Cluster

```sh
cockroach demo --insecure --no-example-database
```

Tables

```sql
CREATE TABLE a (
  "value" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "range_timestamp" TIMESTAMPTZ NOT NULL,
  "range_int" INT NOT NULL,
  "range_float" DECIMAL NOT NULL,
  "range_bytes" BYTES NOT NULL,
  "range_point" GEOMETRY NOT NULL
);

CREATE TABLE b (
  "inc" INT NOT NULL PRIMARY KEY,
  "set" STRING NOT NULL,
  "ref" UUID NOT NULL REFERENCES a("value")
);
```

Data

```sh
go run dgs.go gen data \
--config "examples/wide-coverage/config.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 1 \
--batch 10
```

Query

```sql
SELECT
  LEFT(a.value::STRING, 4),
  a.range_timestamp,
  a.range_int,
  a.range_float,
  LEFT(a.range_bytes, 10),
  st_astext(a.range_point),
  b.inc,
  b.set,
  LEFT(b.ref::STRING, 4)
FROM a
JOIN b ON a.value = b.ref
LIMIT 10;
```

Cleanup

```sql
TRUNCATE pet; TRUNCATE person CASCADE;
```