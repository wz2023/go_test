#编译
./protoc --go_out=. -I=./src src/*.proto

read -p "Press any key to continue..." -n1 -s
