#!/bin/bash
pkill -f goraft
rm log.txt goraft
go build goraft.go
x=$1
if [ $# -eq 0 ]
	then
		x=10
fi
for ((i=0;i<=$x-1;i++))
do
	echo "starting node $i"
	./goraft $x $i >> log.txt &
done
echo "please wait while raft protocol is simulated"
sleep 30
echo "ending simulation, edit the sleep statement in raft.sh if you'd like the simulation to run longer"
echo "see log in log.txt"
pkill -f goraft
echo "simulation complete" >> log.txt
