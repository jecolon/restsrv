#!/bin/sh

for i in `seq 1 10`; do
	sqlite3 -echo posts.db "insert into post(id, userId, title, body)
	values(${i}, 1, 'Post ${i}', 'El contenido del post ${i}.');"
done
