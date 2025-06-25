@echo off
:: 编译proto文件
protoc --go_out=. -I=./src src\*.proto

:: 等待用户按键
echo Press any key to continue...
pause >nul