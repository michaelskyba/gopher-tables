# Gopher Tables
Gopher Tables is an online multiplayer "race" game. Players join a game 
together, and then they race to answer multiplication exercises as fast as 
possible.

The backend is written in Go, using the ``net/http`` package. There's no
frontend framework, only [sakura.css](https://github.com/oxalorg/sakura) for
styling. Gopher Tables was created as a school project, meant for learning Go.

## Database setup
```sh
ssu -s # Any privilege elevation utility would work, e.g. su
pacman -S mariadb # or whatever package manager you have
mariadb-install-db --user=mysql --basedir=/usr --datadir=/var/lib/mysql
systemctl start mariadb # systemd lol
mysql -u root -p # Just press Enter, no password
```
Now you're inside MariaDB:
```
create database db;
use db;
create user 'michael'@'localhost' identified by 'password';
grant all privileges on db.* to 'michael'@'localhost';
flush privileges;
quit
```
Create the tables:
```sh
git clone https://github.com/michaelskyba/gopher-tables
cd gopher-tables
mysql -u michael -p # use "password" as the password
```
```
source db/create_tables.sql
select * from accounts;
quit
```

## Running
```sh
cd /path/to/gopher-tables
go build app.go
./app &
$BROWSER localhost:8000
```

## Primary resources used
- https://learnxinyminutes.com/docs/go/
- https://tour.golang.org/
- https://go.dev/doc/articles/wiki/
- https://go.dev/doc/tutorial/database-access
- https://pkg.go.dev/
