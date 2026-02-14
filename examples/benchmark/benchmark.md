### Setup

Create cluster

```sh
make cluster
```

Create table

```sql
CREATE TABLE member (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING NOT NULL,
  registered TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE member SPLIT AT
  SELECT rpad(to_hex(prefix::INT), 32, '0')::UUID
  FROM generate_series(0, 16) AS prefix;
```

**Wait for leaseholder counts to converge**

### Benchmark

Build

```sh
go build dgs.go
```

Config

```sh
./dgs gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public \
--row-count member:1000000 \
> examples/benchmark/config.yaml
```

Data

```sh
time ./dgs gen data \
--config examples/benchmark/config.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--batch 10000 \
--workers 4
```

Results

| Run | Rows inserted | Time taken |
| --- | ------------- | ---------- |
| 1   | 1,000,000     | 21.084s    |
| 2   | 1,000,000     | 27.788s    |
| 3   | 1,000,000     | 33.428s    |

Queries

```sql
-- Validate correct data count.
SELECT COUNT(*) FROM member; SELECT COUNT(*) FROM product; SELECT COUNT(*) FROM purchase; SELECT COUNT(*) FROM purchase_line;

-- Validate relationships.
SELECT
  m.email,
  p.amount,
  p.status,
  pr.name,
  pl.quantity,
  p.ts
FROM purchase_line pl
JOIN purchase p ON pl.purchase_id = p.id
JOIN product pr ON pl.product_id = pr.id
JOIN member m ON p.member_id = m.id
LIMIT 10;

-- Truncate for the next test.
TRUNCATE purchase_line; TRUNCATE purchase CASCADE; TRUNCATE product CASCADE; TRUNCATE member CASCADE;
```

### Teardown

```sh
roachprod destroy $CLUSTER_NAME
```