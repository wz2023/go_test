cd landlords
go build
del /Q logs
start landlords.exe -log_dir="./logs"  
cd ../hallserver
go build
del /Q logs
start hallserver.exe -log_dir="./logs" 
cd ../fish
go build
del /Q logs
start fish.exe -log_dir="./logs" 
cd ../threecards
go build
del /Q logs
start threecards.exe -log_dir="./logs" 
cd ../ox100
go build
del /Q logs
start ox100.exe -log_dir="./logs" 
cd ../redblack
go build
del /Q logs
start redblack.exe -log_dir="./logs" 
cd ../ox5
go build
del /Q logs
start ox5.exe -log_dir="./logs" 
cd ../dragontiger
go build
del /Q logs
start dragontiger.exe -log_dir="./logs" 
cd ../gateserver
go build
del /Q logs
start gateserver.exe -log_dir="./logs" 