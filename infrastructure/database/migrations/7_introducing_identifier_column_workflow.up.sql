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
  "workflows" as w;

DROP TABLE "workflows";

ALTER TABLE "workflows_tmp"
RENAME TO "workflows";

DROP TABLE workflow_meta;
