consul agent -dev

go run server/server.go "node 1" :5000 localhost:8500

go run server/server.go "node 2" :5001 localhost:8500
