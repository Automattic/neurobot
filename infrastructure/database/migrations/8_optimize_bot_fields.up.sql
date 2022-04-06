create table bots_up
(
    id          integer not null primary key autoincrement,
    description TEXT,
    username    TEXT unique,
    password    TEXT,
    active      integer default 1
);

drop table bots;
alter table bots_up rename to bots;
