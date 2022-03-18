CREATE TABLE "tmp_triggers" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"variety" TEXT,
"workflow_ids" TEXT,
"active" INTEGER DEFAULT 1
);

INSERT INTO "tmp_triggers" (id,name,description,variety,workflow_ids,active)
SELECT
  id,
  name,
  description,
  variety,
  CAST(workflow_id as TEXT),
  active
FROM
  "triggers";

DROP TABLE "triggers";

ALTER TABLE "tmp_triggers"
RENAME TO "triggers";
