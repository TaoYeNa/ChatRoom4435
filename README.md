# ChatRoom4435  
# Require grpc, protoc, consul  
First, run **consul agent -dev**,   
Use make file: **make dep**, **make**  
Run the following commands in three different terminals:  
**go run server/server.go "Node 1" :5000 localhost:8500**  
**go run server/server.go "Node 2" :5001 localhost:8500**  
**go run server/server.go "Node 3" :5002 localhost:8500**  
