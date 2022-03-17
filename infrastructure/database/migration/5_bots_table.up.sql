CREATE TABLE "bots" (
"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
"identifier" TEXT UNIQUE,
"name" TEXT,
"description" TEXT,
"username" TEXT,
"password" TEXT,
"created_by" TEXT,
"active" integer DEFAULT 1
);
