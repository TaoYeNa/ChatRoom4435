# ChatRoom4435  
## Require grpc, protoc, consul  
First, run **consul agent -dev**,   
Use make file: **make dep**, **make**  
Run the following commands in three different terminals:  
**go run server/server.go "Node 1" :5000 localhost:8500**  
**go run server/server.go "Node 2" :5001 localhost:8500**  
**go run server/server.go "Node 3" :5002 localhost:8500**  
Open template/assest/html/index.html in the web browser, you can open multiple window if you want.
![image](https://github.com/TaoYeNa/ChatRoom4435/blob/fac8e4b1ea5a61288ee321eccc2fe38a7d91cb77/%E6%88%AA%E5%B1%8F2022-04-14%20%E4%B8%8B%E5%8D%888.49.10.png)
