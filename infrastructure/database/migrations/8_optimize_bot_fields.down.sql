create table bots_down
(
    id          integer not null primary key autoincrement,
    identifier  TEXT unique,
    name        TEXT,
    description TEXT,
    username    TEXT,
    password    TEXT,
    created_by  TEXT,
    active      integer default 1
);

drop table bots;
alter table bots_down rename to bots;
