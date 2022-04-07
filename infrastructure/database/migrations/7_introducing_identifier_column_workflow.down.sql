CREATE TABLE "workflow_meta" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"workflow_id" integer,
"key" TEXT,
"value" TEXT
);

INSERT INTO workflow_meta ('workflow_id','key','value')
SELECT id,'toml_identifier',identifier FROM workflows as w where identifier <> "" w AND w.id != 1

CREATE TABLE "workflows_tmp" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"active" integer DEFAULT 1
);

INSERT INTO "workflows_tmp" as t (id,name,description,active)
SELECT
  id,
  name,
  description,
  active
FROM
  "workflows";

DROP TABLE "workflows";

ALTER TABLE "workflows_tmp"
RENAME TO "workflows";
