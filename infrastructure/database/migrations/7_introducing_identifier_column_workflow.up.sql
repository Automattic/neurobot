CREATE TABLE "workflows_tmp" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"active" integer DEFAULT 1,
"identifier" TEXT UNIQUE
);

INSERT INTO "workflows_tmp" as t (id,name,description,active,identifier)
SELECT
  id,
  name,
  description,
  active,
  (SELECT value FROM "workflow_meta" WHERE key = 'toml_identifier' AND workflow_id = w.id)
FROM
  "workflows" as w
WHERE w.id != 1;

/* quickstart demo doesn't have any identifier rows in workflow_meta and above query fails to convert NULL to string (SQLite) */
/* so we exclude that workflow in the above query and do an insert using the query below */
INSERT INTO "workflows_tmp" as t (id,name,description,active,identifier)
SELECT
  id,
  name,
  description,
  active,
  "quickstart"
FROM
  "workflows" as w
WHERE w.id = 1;

DROP TABLE "workflows";

ALTER TABLE "workflows_tmp"
RENAME TO "workflows";

DROP TABLE workflow_meta;
