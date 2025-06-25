#!/bin/bash
#git pull

rm -f version/version.go
git rev-list HEAD | sort > config.git-hash
LOCALVER=`wc -l config.git-hash | awk '{print $1}'`
if [ $LOCALVER \> 1 ] ; then
    VER=`git rev-list origin/master | sort | join config.git-hash - | wc -l | awk '{print $1}'`
    if [ $VER != $LOCALVER ] ; then
        VER="$VER+$(($LOCALVER-$VER))"
    fi
    if git status | grep -q "modified:" ; then
        VER="${VER}M"
    fi
    VER="$VER:$(git rev-list HEAD -n 1 | cut -c 1-7)"
    GIT_VERSION=r$VER
else
    GIT_VERSION=
    VER="x"
fi
rm -f config.git-hash

cat version/version.go.tmp | sed "s/\$FULL_VERSION/$GIT_VERSION/g" > version/version.go

echo "Generated version.go"
go build
echo "build blackjack"

root_path=E:/workspace/fish_game/newstars
bin_path=$root_path/bin/blackjackServer/
bin_path2=$root_path/bin/

if [ ! -d $bin_path ];then
      mkdir -p $bin_path
      mkdir -p $bin_path/conf
      mkdir -p $bin_path/logs
fi

cp $root_path/Server/blackjack/blackjack $bin_path2
mv $root_path/Server/blackjack/blackjack $bin_path

cp $root_path/Server/blackjack/config.yaml $bin_path
cp $root_path/Server/blackjack/conf/conf.json $bin_path/conf/
cp $root_path/Server/blackjack/conf/blacklist.json $bin_path/conf/
cp $root_path/Server/blackjack/conf/17monipdb.dat $bin_path/conf/
echo "cp blackjack to bin path"

