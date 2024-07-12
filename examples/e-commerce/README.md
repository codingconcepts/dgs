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

Data

```sh
go run dgs.go gen data \
--config "examples/e-commerce/config.yaml" \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 4 \
--batch 10000
```

Query

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

Cleanup

```sql
TRUNCATE purchase_line; TRUNCATE purchase CASCADE; TRUNCATE product CASCADE; TRUNCATE member CASCADE;
```

Config

```sh
go run dgs.go gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public
```