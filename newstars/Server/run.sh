supervisorctl stop all

cd landlords
go build
supervisorctl start landlords

cd ../threecards
go build
supervisorctl start threecards

cd ../ox5
go build
supervisorctl start ox5

cd ../ox100
go build
supervisorctl start ox100

cd ../dragontiger
go build
supervisorctl start dragontiger

cd ../redblack
go build
supervisorctl start redblack

cd ../benz
go build
supervisorctl start benz

cd ../baccarat
go build
supervisorctl start baccarat

cd ../fish
go build
supervisorctl start fish

cd ../payserver
go build
supervisorctl start payserver

# cd ../gamechannel
# go build
# # supervisorctl start gamechannel

cd ../hallserver
go build
supervisorctl start hallserver


# cd ../markeserver
# go build
# supervisorctl start markeserver


cd ../wagency
go build
supervisorctl start wagency

cd ../gateserver
go build
supervisorctl start gateserver

