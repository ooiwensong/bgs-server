# BGS-Server

BGS-Server is a RESTful API backend Go web server for my [Boardgame Scheduler](https://github.com/ooiwensong/boardgame-scheduler?tab=readme-ov-file) project. In the original Boardgame Scheduler repository, the backend server was written in Node.js/ Express. As a way to practice Go, I have decided to re-write the entire backend server in the language.

The server uses the `chi` package as the primary router for HTTP services. It is lightweight, idiomatic package that is built on top of the `net/http` package of the Go standard library. Other packages used include `pq` as the driver to PostgreSQL databases, `golang-jwt` package to implement jwt authentication, as well as the `crypto` package from golang.org.
