#!/bin/bash

if [ "xxx$1" == "xxx" ] ; then
	echo "usage: $0 [options]"
	echo ""
	echo "examples:"
	echo "$0 -v MySQLs Production   #all production mysqls"
	echo "$0 -v Memcached Staging   #all staging memcacheds"
	echo "$0 -v SF MySQLs           #all MySQLs in SF"
fi

#( echo [DyamicPool] && arsimto ls -i $1 $2 $3 $4 $5 -d=name,ip | awk '{print $2"  assetname="$1}' ) > inventory/dynamic.cfg
( echo [DyamicPool] && arsimto ls -i $1 $2 $3 $4 $5 -d=name,ip | fgrep -v + | awk '{print $2"	assetname="$1}' )

