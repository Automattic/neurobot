CREATE TABLE "room_members" (
"user_id"  TEXT NOT NULL CHECK(user_id <> ''),
"room_id"  TEXT NOT NULL CHECK(room_id <> '')
);
