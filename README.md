# Gopher Tables
Gopher Tables is an online multiplayer "race" game. Players join a game 
together, and then they race to answer multiplication exercises as fast as 
possible.

The backend is written in Go, using the ``net/http`` package. There's no
frontend framework, only [sakura.css](https://github.com/oxalorg/sakura) for
styling. Gopher Tables was created as a school project, meant for learning Go.

## Running the backend
```sh
git clone https://github.com/michaelskyba/gopher-tables
cd gopher-tables
go build app.go
./app &
$BROWSER localhost:8000
```

## Primary resources used
- https://learnxinyminutes.com/docs/go/
- https://tour.golang.org/
- https://go.dev/doc/articles/wiki/
- https://pkg.go.dev/
