comp: 
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=require_unimplemented_servers=false:. \
	--go-grpc_opt=paths=source_relative proto/ChatRoom4435.proto

dep:
	go mod init ChatRoom4435
	go mod tidy

clean:
	rm go.sum
	rm go.mod	
	 	
