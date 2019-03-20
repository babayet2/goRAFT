package main
//package server

//import "errors"
//import "log"
import (
	"net"
	"net/rpc"
	"net/http"
	"fmt"
	"time"
	"strconv"
	"log"
	"os"
	"math/rand"
)

type State int

const (
	follower State = 0
	candidate State = 1
	leader State = 2
)

//store node information
type Node struct {
	nodes int
	pid int
	term int
	election_timeout time.Duration
	last_com time.Time
	state State
	//256 is currently the max number of nodes
	clients [256]*rpc.Client
}

func (n *Node) log(str string){
	fmt.Printf("[PID %03d | TERM %03d]  %s\n", n.pid, n.term, str);
}

//does not have to be exported
func (n *Node) init_channels(){
	for i:=0; i<n.nodes; i++ {
		//attempt to initiate connections with each client
		client, err := rpc.DialHTTP("tcp", "127.0.0.1:" + strconv.Itoa(1200 + i))
		if err != nil {
			//retry indefinitely if connection fails
			//log.Fatal("dialing erorr, will retry:", err)
			for err!=nil{
				client, err = rpc.DialHTTP("tcp", "127.0.0.1:" + strconv.Itoa(1200 + i))
			}
		}
		n.clients[i] = client
	}
	return
}

//does not have to be exported
func (n *Node) init_server(){
	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":" + strconv.Itoa(1200 + n.pid))
	if err != nil {
		n.log("listen error")
	}
	n.log("Serving RPC server on port: " + strconv.Itoa(1200 + n.pid))
	// Start accept incoming HTTP connections
	go http.Serve(l, nil)

}

func (n *Node) Receive_vote_request(term int, reply *int) error{
	//fmt.Println(term, n.term)
	n.last_com = time.Now()
	if term>n.term{
		if n.state==leader{
			n.state = follower
		}
		n.term = term
		*reply = 1
		return nil
	}
	*reply = 0
	return nil
}

func (n *Node) Receive_heartbeat(term int, reply *int) error{
	n.last_com = time.Now()
	if term>n.term{
		n.term = term
		if n.state==leader{
			n.state = follower
		}
	}
	*reply = 0
	return nil
}

func (n *Node) request_vote(pid int) int{
	var reply int
	err := n.clients[pid].Call("Node.Receive_vote_request", n.term, &reply)
	if err!=nil{
		//log.Fatal("vote request:", err)
	}
	return reply
}

func (n *Node) send_heartbeat(pid int) int{
	var reply int
	err := n.clients[pid].Call("Node.Receive_heartbeat", n.term, &reply)
	if err!=nil{
		//log.Fatal("heartbeat request:", err)
	}
	return reply
}

func (n *Node) send_heartbeats(){
	for i:=0;i<n.nodes;i++{
		go n.send_heartbeat(i)
	}
}

func (n *Node) leader_crash(){
	time.Sleep(time.Second * 6)
	n.log("simulated crash, killing process");
	//log.Fatal("[PID " + strconv.Itoa(n.pid) + "] simulated crash, killing process");
	os.Exit(0)
}

func (n *Node) Receive_leader_elected(term int, reply *int) error{
	if term>n.term{
		n.term = term
		n.state = follower
		n.last_com = time.Now()
	}
	*reply = 0
	return nil
}

func (n *Node) broadcast_election(){
	var reply int
	for i:=0;i<n.nodes;i++{
		if i==n.pid{
			continue
		}
		go n.clients[i].Call("Node.Receive_leader_elected", n.term, &reply)
	}
}

func (n *Node) state_loop(){
	//inifite loop
	for{
		if n.state==follower{
			//loop in follower state unless election times out
			for (n.election_timeout > (time.Now().Sub(n.last_com))){
			}
			//fmt.Println(n.election_timeout, (time.Now().Sub(n.last_com)))
			//the election timer has timed out!
			n.state = candidate
		}
		if n.state==candidate{
			sum := 1
			n.term += 1
			n.log("election timeout! becoming candidate")
			stored_term := n.term
			for i:=0;i<n.nodes;i++{
				sum += n.request_vote(i)
			}
			if(n.term==stored_term){
				n.log("recieved " + strconv.Itoa(sum) + " votes")
				if sum>n.nodes/2{
					n.state = leader
					go n.leader_crash()
					n.log("recieved vote majority! becoming leader")
					n.broadcast_election()
					continue
				}
			}
			n.state = follower
		}
		if n.state==leader{
			n.log("broadcasting heartbeats")	
			n.send_heartbeats()
			n.last_com = time.Now()
			time.Sleep(time.Millisecond * 30)
		}
	}
}

func main(){
	n := new(Node)
	
	i, err := strconv.Atoi(os.Args[1])
	n.nodes = i
	if err!=nil{ 
		log.Fatal("error parsing command line arguments", err)
	}
	i, err = strconv.Atoi(os.Args[2])
	n.pid = i
	if err!=nil {
		log.Fatal("error parsing command line arguments", err)
	}
	n.init_server()
	n.init_channels()
	n.log("sucessfully initialized server and connected to all channels")
	rand.Seed(time.Now().UnixNano())
	n.last_com = time.Now()
	n.election_timeout = time.Duration(rand.Intn(150) + 150) * time.Millisecond
	//fmt.Println("[PID " + strconv.Itoa(n.pid) + "] election timeout: ", n.election_timeout)
	n.state_loop()
	
	//stop server from closing immediately
	var input string
	fmt.Scanln(&input)
	
}
