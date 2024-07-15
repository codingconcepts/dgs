Cluster

```sh
cockroach demo --insecure --no-example-database
```

Tables

```sql
CREATE TABLE person (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name STRING NOT NULL
);

CREATE TABLE pet (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name STRING NOT NULL,
  owner_id UUID NOT NULL REFERENCES person(id)
);
```

Config

```sh
go run dgs.go gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public \
--row-count person:1000 \
--row-count pet:100000 \
> examples/simple-relational/config.yaml
```

Data

```sh
go run dgs.go gen data \
--config "examples/simple-relational/config.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 1 \
--batch 10000
```

Query

```sql
SELECT
  pr.id,
  pt.owner_id
FROM person pr
JOIN pet pt ON pr.id = pt.owner_id
ORDER BY pr.id
LIMIT 10;

SELECT p.owner_id, COUNT(*)
FROM pet p
GROUP BY p.owner_id
ORDER BY 2 DESC;
```

Cleanup

```sql
TRUNCATE pet; TRUNCATE person CASCADE;
```