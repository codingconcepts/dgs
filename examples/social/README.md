Cluster

```sh
cockroach demo --insecure --no-example-database
```

Tables

```sql
CREATE TABLE member (
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
