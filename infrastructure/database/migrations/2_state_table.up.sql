CREATE TABLE "matrix_state" (
"bot_id" INTEGER NOT NULL CHECK (bot_id <> ''),
"what"   TEXT NOT NULL CHECK(what <> ''),
"id"     TEXT NOT NULL CHECK(id <> ''),
"value"  TEXT NOT NULL
);
