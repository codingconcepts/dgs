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
```

Config (with default row counts)

```sh
go run dgs.go gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public > examples/e-commerce/config.yaml
```

Config (with custom row counts)

```sh
go run dgs.go gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public \
--row-count member:100000 \
--row-count product:10000 \
--row-count purchase:200000 \
--row-count purchase_line:400000 > examples/e-commerce/config.yaml
```

Data

```sh
go run dgs.go gen data \
--config "examples/e-commerce/config.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 4 \
--batch 10000 \
--cpu-profile cpuprof
```

Profile

```sh
go tool pprof cpu.pprof

(pprof) top

go tool pprof -png cpu.pprof > cpupprof.png
```

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