CREATE TABLE "triggers" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"variety" TEXT,
"workflow_id" INTEGER NOT NULL,
"active" INTEGER DEFAULT 1
);

CREATE TABLE "trigger_meta" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"trigger_id" integer,
"key" TEXT,
"value" TEXT
);
