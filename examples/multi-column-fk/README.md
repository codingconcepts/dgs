```sh
cockroach demo --insecure --no-example-database
```

Create tables

```sql
CREATE TABLE public.users (
  id UUID NOT NULL,
  city VARCHAR NOT NULL,
  CONSTRAINT users_pkey PRIMARY KEY (city ASC, id ASC)
);

CREATE TABLE public.user_promo_codes (
  city VARCHAR NOT NULL,
  user_id UUID NOT NULL,
  code VARCHAR NOT NULL,
  CONSTRAINT user_promo_codes_pkey PRIMARY KEY (city ASC, user_id ASC, code ASC),
  CONSTRAINT user_promo_codes_city_user_id_fkey FOREIGN KEY (city, user_id) REFERENCES public.users(city, id)
);
```

Generate data

```sh
go run dgs.go gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public \
--row-count users:1 \
--row-count user_promo_codes:1 \
> dgs.yaml

go run dgs.go gen data \
--workers 1 \
--batch 1 \
--url "postgres://root@localhost:26257?sslmode=disable" \
--config dgs.yaml
```

promo_codes                - 1000
rides                      -  500
user_promo_codes           -    5
users                      -   50
vehicles                   -   15
vehicle_location_histories - 1000

### Debugging

Existing column statement

```sql
WITH
  columns_info AS (
    SELECT 
      table_name,
      column_name,
      ordinal_position,
      column_default,
      is_nullable,
      character_maximum_length,
      udt_name AS data_type
    FROM information_schema.columns
    WHERE table_schema = 'public'
  ),
  foreign_keys_info AS (
    SELECT
      tc.constraint_name,
      tc.table_name AS fk_table,
      kcu.column_name AS fk_column,
      ccu.table_name AS pk_table,
      ccu.column_name AS pk_column
    FROM information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
  ),
  user_defined_types AS (
    SELECT 
      t.typname AS type_name,
      t.typtype AS type_type,
      array_agg(e.enumlabel ORDER BY e.enumsortorder) AS enum_labels
    FROM  pg_type t
    JOIN pg_namespace n ON t.typnamespace = n.oid
    LEFT JOIN pg_enum e ON t.oid = e.enumtypid
    WHERE n.nspname = 'public'
    GROUP BY t.typname, t.typtype
  )
SELECT
  fk.constraint_name,
  c.table_name,
  c.column_name,
  c.column_default,
  c.is_nullable,
  c.character_maximum_length,
  CASE
    WHEN udt.type_type = 'e' THEN 'enum'
    ELSE c.data_type
  END AS "data_type",
  CASE 
    WHEN udt.type_type = 'e' THEN udt.enum_labels
    ELSE NULL
  END AS user_defined_type,
  CASE
    WHEN fk.pk_table IS NOT NULL THEN fk.pk_table || '.' || fk.pk_column
  END AS "fk"
FROM columns_info AS c
LEFT JOIN foreign_keys_info AS fk
ON c.table_name = fk.fk_table AND c.column_name = fk.fk_column
LEFT JOIN user_defined_types AS udt
ON c.data_type = udt.type_name
ORDER BY c.table_name, c.ordinal_position;
```

New column statement

```sql
WITH
  columns_info AS (
    SELECT 
      table_name,
      column_name,
      ordinal_position,
      column_default,
      is_nullable,
      character_maximum_length,
      udt_name AS data_type
    FROM information_schema.columns
    WHERE table_schema = 'public'
  ),
  foreign_keys_info AS (
    SELECT
      tc.constraint_name,
      tc.table_name AS fk_table,
      kcu.column_name AS fk_column,
      ccu.table_name AS pk_table,
      ccu.column_name AS pk_column
    FROM information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
  ),
  user_defined_types AS (
    SELECT 
      t.typname AS type_name,
      t.typtype AS type_type,
      array_agg(e.enumlabel ORDER BY e.enumsortorder) AS enum_labels
    FROM  pg_type t
    JOIN pg_namespace n ON t.typnamespace = n.oid
    LEFT JOIN pg_enum e ON t.oid = e.enumtypid
    WHERE n.nspname = 'public'
    GROUP BY t.typname, t.typtype
  )
SELECT
  c.table_name,
  c.column_name,
  c.column_default,
  c.is_nullable,
  c.character_maximum_length,
  CASE
    WHEN udt.type_type = 'e' THEN 'enum'
    ELSE c.data_type
  END AS "data_type",
  CASE 
    WHEN udt.type_type = 'e' THEN udt.enum_labels
    ELSE NULL
  END AS user_defined_type,
  array_agg(DISTINCT fk.pk_table || '.' || fk.pk_column) FILTER (WHERE fk.pk_table IS NOT NULL) AS "fk"
FROM columns_info AS c
LEFT JOIN foreign_keys_info AS fk
ON c.table_name = fk.fk_table AND c.column_name = fk.fk_column
LEFT JOIN user_defined_types AS udt
ON c.data_type = udt.type_name
GROUP BY 
  c.table_name,
  c.column_name,
  c.ordinal_position,
  c.column_default,
  c.is_nullable,
  c.character_maximum_length,
  c.data_type,
  udt.type_type,
  udt.enum_labels
ORDER BY c.table_name, c.ordinal_position;
```