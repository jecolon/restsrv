#!/bin/sh

sqlite3 posts.db "drop table if exists post;
create table if not exists post(
	id int primary key not null,
	userId int not null,
	title text not null,
	body text not null
);"
