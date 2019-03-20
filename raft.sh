#!/bin/bash
pkill -f goraft > /dev/null
rm log.txt goraft
go build goraft.go
x=$1
if [ $# -eq 0 ]
	then
		x=10
fi
for ((i=0;i<$x;i++))
do
	echo "starting node $i"
	./goraft $x $i >> log.txt &
done
echo "please wait while raft protocol is simulated"
sleep 30
echo "ending simulation, edit the sleep statement in raft.sh if you'd like the simulation to run longer"
echo "see log in log.txt"

#kill all goraft processes, and surpress error output from pkill
exec 3>&2
exec 2> /dev/null
pkill -f goraft
exec 2>&3

echo "simulation complete" >> log.txt
