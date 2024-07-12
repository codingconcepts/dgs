WITH
  columns_info AS (
    SELECT 
      table_name,
      column_name,
      ordinal_position,
      column_default,
      is_nullable,
      udt_name AS data_type
    FROM information_schema.columns
    WHERE table_schema = $1
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
    AND tc.table_schema = $1
  ),
  user_defined_types AS (
    SELECT 
      t.typname AS type_name,
      t.typtype AS type_type,
      array_agg(e.enumlabel ORDER BY e.enumsortorder) AS enum_labels
    FROM  pg_type t
    JOIN pg_namespace n ON t.typnamespace = n.oid
    LEFT JOIN pg_enum e ON t.oid = e.enumtypid
    WHERE n.nspname = $1
    GROUP BY t.typname, t.typtype
  )
SELECT 
  c.table_name,
  c.column_name,
  c.column_default,
  c.is_nullable,
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
ORDER BY c.table_name, c.ordinal_position