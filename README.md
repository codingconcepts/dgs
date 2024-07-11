# dgs
A streaming version of dg, which writes data directly to a database without any kind of buffering.

### Local example

Cluster

```sh
cockroach demo --insecure --no-example-database
```

Tables

```sql
CREATE TABLE member (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING NOT NULL,
  registered TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE product (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name STRING NOT NULL,
  price DECIMAL NOT NULL
);

CREATE TYPE purchase_status AS ENUM ('pending', 'paid', 'dispatched');
CREATE TABLE purchase (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  member_id UUID NOT NULL REFERENCES member(id),
  amount DECIMAL NOT NULL,
  status purchase_status NOT NULL ,
  ts TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE purchase_line (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  purchase_id UUID NOT NULL REFERENCES purchase(id),
  product_id UUID NOT NULL REFERENCES product(id),
  quantity INT NOT NULL DEFAULT 1
);

CREATE TABLE person (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING NOT NULL,
  full_name STRING NOT NULL,
  registered TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pet (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  person_id UUID NOT NULL REFERENCES person(id),
  name STRING NOT NULL,
  registered TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE person SPLIT AT
  SELECT rpad(to_hex(prefix::INT), 32, '0')::UUID AS split_at
  FROM generate_series(0, 16) AS prefix;

ALTER TABLE pet SPLIT AT
  SELECT rpad(to_hex(prefix::INT), 32, '0')::UUID AS split_at
  FROM generate_series(0, 16) AS prefix;
```

### Generate data

```sh
time go run dgs.go \
--config "examples/stress.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 4 \
--batch 10000
# 11.747s


```

### Scratchpad

Select referenced data

```sql
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
```

Get data size

```sql
SELECT
  range_id,
  ROUND(range_size_mb) AS range_size_mb,
  span_stats->'key_count' AS row_count
FROM [SHOW RANGES FROM TABLE member WITH DETAILS];

SELECT
  range_id,
  ROUND(range_size_mb) AS range_size_mb,
  span_stats->'key_count' AS row_count
FROM [SHOW RANGES FROM TABLE product WITH DETAILS];

SELECT
  range_id,
  ROUND(range_size_mb) AS range_size_mb,
  span_stats->'key_count' AS row_count
FROM [SHOW RANGES FROM TABLE purchase WITH DETAILS];

SELECT
  *
FROM [SHOW RANGES FROM TABLE purchase_line with INDEXES, KEYS, DETAILS];
```

Truncate tables

```sql
TRUNCATE TABLE purchase_line; TRUNCATE TABLE purchase CASCADE; TRUNCATE TABLE product CASCADE; TRUNCATE TABLE member CASCADE;
```

### Todo

- [ ] [Bug] Add length field to range (to prevent Int63n from failing because of max - min = 0 error)

- [ ] [Performance] Process ref dependency tables first and run them concurrently
- [ ] [Performance] Run inserts in parallel
- [ ] [Performance] Use ints for min and max ranges where possible
- [ ] [Performance] Consider sorting data by primary key column(s) before inserting
