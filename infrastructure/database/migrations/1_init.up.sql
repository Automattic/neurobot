CREATE TABLE "bots" (
"id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
"description" TEXT,
"username"    TEXT UNIQUE,
"password"    TEXT,
"active"      INTEGER DEFAULT 1
);

CREATE TABLE "workflow_step_meta" (
"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
"step_id" INTEGER,
"key" TEXT,
"value" TEXT
);

CREATE TABLE "workflow_steps" (
"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"variety" TEXT,
"workflow_id" INTEGER,
"sort_order" INTEGER,
"active" INTEGER DEFAULT 1
);

CREATE TABLE "workflows" (
"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
"name" TEXT,
"description" TEXT,
"active" INTEGER DEFAULT 1,
"identifier" TEXT UNIQUE
);
