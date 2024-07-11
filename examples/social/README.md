Cluster

```sh
cockroach demo --insecure --no-example-database
```

Tables

```sql
CREATE TABLE users (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" STRING NOT NULL,
  "full_name" STRING NOT NULL,
  "location" GEOMETRY NOT NULL
);
```

Data

```sh
go run dgs.go \
--config "examples/social/config.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 4 \
--batch 10000
```

Query

```sql
SELECT
  "id",
  "email",
  "full_name",
  st_astext("location")
FROM users
WHERE st_dwithin(
  "location",
  ST_GeomFromText('POINT(-0.8911379425829405 51.04284752235447)'),
  10
)
LIMIT 10;
```