


cd landlords
scp landlords root@47.244.140.195:/opt/newstars/bin/landlords/


cd ../threecards
scp threecards root@47.244.140.195:/opt/newstars/bin/threecards/


cd ../ox5
scp ox5 root@47.244.140.195:/opt/newstars/bin/ox5/


cd ../ox100
scp ox100 root@47.244.140.195:/opt/newstars/bin/ox100/


cd ../dragontiger
scp dragontiger root@47.244.140.195:/opt/newstars/bin/dragontiger/


cd ../redblack
scp redblack root@47.244.140.195:/opt/newstars/bin/redblack/


cd ../benz
scp benz root@47.244.140.195:/opt/newstars/bin/benz/

cd ../baccarat
scp baccarat root@47.244.140.195:/opt/newstars/bin/baccarat/

cd ../fish
scp fish root@47.244.140.195:/opt/newstars/bin/fish/

cd ../hallserver
scp hallserver root@47.244.140.195:/opt/newstars/bin/hallserver/

# cd ../payserver
# scp payserver root@47.244.140.195:/opt/newstars/bin/payserver/

 cd ../wagency
scp wagency root@47.244.140.195:/opt/newstars/bin/wagency/

cd ../gateserver
scp gateserver root@47.244.140.195:/opt/newstars/bin/gateserver/

# supervisorctl stop 	newbaccarat
# supervisorctl stop 	newbenz
# supervisorctl stop 	newdragontiger
# supervisorctl stop 	newfish
# supervisorctl stop 	newhallserver
# supervisorctl stop 	newlandlords
# supervisorctl stop 	newox100
# supervisorctl stop 	newox5
# supervisorctl stop 	newredblack
# supervisorctl stop 	newthreecards
# supervisorctl stop 	wagency
# supervisorctl stop 	newgateserver

# supervisorctl start newbaccarat
# supervisorctl start newbenz
# supervisorctl start newdragontiger
# supervisorctl start newfish
# supervisorctl start newhallserver
# supervisorctl start newlandlords
# supervisorctl start newox100
# supervisorctl start newox5
# supervisorctl start newredblack
# supervisorctl start newthreecards
# supervisorctl start wagency
# supervisorctl start newgateserver

#scp 0606.tar.gz root@47.244.140.195:/opt/newstars/bin/
# supervisorctl stop all
# supervisorctl start all
#tar -czvf 0606.tar.gz gateserver/gateserver  hallserver/hallserver fish/fish ox5/ox5 threecards/threecards landlords/landlords  benz/benz baccarat/baccarat redblack/redblack dragontiger/dragontiger ox100/ox100 wagency/wagency
# tar -xzvf 0606.tar.gz
