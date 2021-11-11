CREATE TABLE "workflows" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"active" integer DEFAULT 1
);

CREATE TABLE "triggers" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"variety" TEXT,
"workflow_ids" TEXT,
"active" INTEGER DEFAULT 1
);

CREATE TABLE "workflow_steps" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"variety" TEXT,
"workflow_id" integer,
"sort_order" integer,
"active" INTEGER DEFAULT 1
);

CREATE TABLE "trigger_meta" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"trigger_id" integer,
"key" TEXT,
"value" TEXT
);

CREATE TABLE "workflow_step_meta" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"step_id" integer,
"key" TEXT,
"value" TEXT
);
