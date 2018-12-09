package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"

	m "../lib/message"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
)

type Status int

type BrokerRPCServer int

type consumerId string

type record [512]byte

type partition []*record

type Peer struct {
	addr net.Addr
}

type Topic struct {
	topicID        string
	partitionIdx   uint8
	partition      partition
	consumerOffset map[consumerId]uint
	Status
	FollowerList map[net.Addr]bool
}

const (
	Leader Status = iota
	Follower
)

type broker struct {
	topicList map[string]*Topic
}

var b *broker

// Initialize starts the node as a Broker node in the network
func InitBroker(addr string) error {

	b = &broker{
		topicList: make(map[string]*Topic),
	}

	go spawnListener(addr)

	if err := registerBrokerWithManager(); err != nil {
		return err
	}

	fmt.Println("Init Borker")

	for {
	}
	return nil
}

// Spawn a rpc listen client
func spawnListener(addr string) {
	fmt.Println(addr)

	bRPC := new(BrokerRPCServer)
	server := rpc.NewServer()
	server.Register(bRPC)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	fmt.Printf("Serving Server at: %v\n", tcpAddr.String())

	vrpc.ServeRPCConn(server, listener, logger, loggerOptions)
}

// func handleConnection(conn net.Conn) {
// 	dec := gob.NewDecoder(conn)
// 	message := &m.Message{}
// 	dec.Decode(message)

// 	switch message.Type {

// 	case m.NEW_TOPIC:
// 		if err := startLeader(message); err != nil {

// 		} else {
// 			enc := gob.NewEncoder(conn)
// 			if err = enc.Encode(message); err != nil {
// 				log.Fatal("encode error:", err)
// 			}
// 		}

// 	default:
// 		// freebsd, openbsd,
// 		// plan9, windows...
// 	}

// }

// func startBroker(message *m.Message) error {

// 	topic := &Topic{
// 		topicID:        message.Topic,
// 		partitionIdx:   message.Partition,
// 		partition:      partition{},
// 		consumerOffset: make(map[consumerId]uint),
// 		Status:         Leader,
// 		FollowerList:   make(map[net.Addr]bool),
// 	}

// 	if _, exist := b.topicList[topic.topicID]; exist {
// 		return fmt.Errorf("Topic ID has already existed")
// 	}

// 	followersIP := strings.Split(message.Text, ",")

// 	followerMessage := &m.Message{Topic: message.Topic, Partition: message.Partition}

// 	for _, ip := range followersIP {
// 		go broadcastToFollowers(m, ip)
// 	}

// 	b.topicList[topic.topicID] = topic

// 	// b.topicList = append(b.topicList, topic)

// 	fmt.Println("Started Leader")

// 	return nil
// }

// func startLeader(message *m.Message) error {
// 	topic := &Topic{
// 		topicID:        message.Topic,
// 		partitionIdx:   message.Partition,
// 		partition:      partition{},
// 		consumerOffset: make(map[consumerId]uint),
// 		Status:         Leader,
// 		FollowerList:   make(map[net.Addr]bool),
// 	}
// 	return nil

// }

// func broadcastToFollowers(message *m.Message, addr string) error {

// 	destAddr, err := net.ResolveTCPAddr("tcp", addr)
// 	if err != nil {
// 		return err
// 	}

// 	lAddr, err := net.ResolveTCPAddr("tcp", config.BrokerIPPort)
// 	if err != nil {
// 		return err
// 	}

// 	preparePhase(message, destAddr)

// 	conn, err := net.DialTCP("tcp", lAddr, destAddr)

// 	if err != nil {
// 		return err
// 	}

// 	enc := gob.NewEncoder(conn)

// 	if err := enc.Encode(message); err != nil {
// 		return err
// 	}

// 	dec := gob.NewDecoder(conn)

// 	revMessage := &m.Message{}

// 	if err := dec.Decode(revMessage); err != nil{
// 		return err
// 	}

// 	return nil
// }

func (bs *BrokerRPCServer) StartLeader(message *m.Message, ack *bool) error {
	*ack = true

	fmt.Println("Leader Topic Name", message.Topic)

	topic := &Topic{
		topicID:        message.Topic,
		partitionIdx:   message.Partition,
		partition:      partition{},
		consumerOffset: make(map[consumerId]uint),
		Status:         Leader,
		FollowerList:   make(map[net.Addr]bool),
	}

	if _, exist := b.topicList[topic.topicID]; exist {
		return fmt.Errorf("Topic ID has already existed")
	}

	if message.Text == "" {
		fmt.Println("text is empty")
		return nil
	}

	followersIP := strings.Split(message.Text, ",")

	followerMessage := &m.Message{Topic: message.Topic, Partition: message.Partition}

	fmt.Printf("%+v\n", followerMessage)

	var waitGroup sync.WaitGroup

	waitGroup.Add(len(followersIP))

	fmt.Println("follower len", len(followersIP))

	for _, ip := range followersIP {
		fmt.Println("Looping")
		go broadcastToFollowers(*followerMessage, ip, &waitGroup)
	}
	waitGroup.Wait()

	return nil
}

func broadcastToFollowers(message m.Message, addr string, w *sync.WaitGroup) error {
	destAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	// lAddr, err := net.ResolveTCPAddr("tcp", config.BrokerIP)
	// if err != nil {
	// 	return err
	// }

	conn, err := net.DialTCP("tcp", nil, destAddr)

	if err != nil {
		fmt.Println(err)
	}

	rpcClient := rpc.NewClient(conn)

	var ack bool

	if err := rpcClient.Call("BrokerRPCServer.StartFollower", message, &ack); err != nil {
		return err
	}

	w.Done()

	return nil
}

func (bs *BrokerRPCServer) Ping(message *m.Message, ack *bool) error {
	
	fmt.Println("I've been pinged by: ", message.Text)
	for ;;{
	}
	*ack = true
	return nil
}

func (bs *BrokerRPCServer) StartFollower(message *m.Message, ack *bool) error {
	fmt.Println("Start Follower")
	fmt.Println(message.Topic)

	*ack = true

	topic := &Topic{
		topicID:        message.Topic,
		partitionIdx:   message.Partition,
		partition:      partition{},
		consumerOffset: make(map[consumerId]uint),
		Status:         Leader,
		FollowerList:   make(map[net.Addr]bool),
	}

	if _, exist := b.topicList[topic.topicID]; exist {
		*ack = false
		return fmt.Errorf("Topic ID has already existed")
	}

	return nil
}

// func (b *BrokerServer) InitNewTopic(m *Message, res *bool) error {
// 	topic := new(Topic)
// 	topic.id = m.Topic

// 	Broker.topicList[topic.id] = topic

// 	*res = true
// 	return nil
// }

// func (b *BrokerServer) AppendToPartition(m *Message, res *bool) error {
// 	topicId := m.Topic
// 	var rec record
// 	copy(rec[:], m.Payload.Marshall())
// 	Broker.topicList[topicId].partition = append(Broker.topicList[topicId].partition, &rec)
// 	*res = true
// 	return nil
// }

// func (b *BrokerServer) AddClient(m *Message, res *bool) error {
// 	topicId := m.Topic
// 	var rec record
// 	copy(rec[:], m.Payload.Marshall())
// 	Broker.topicList[topicId].partition = append(Broker.topicList[topicId].partition, &rec)
// 	*res = true
// 	return nil
// }

// func (b *BrokerServer) DispatchData(m *Message, res *bool) error {
// 	topicID := m.Topic
// 	clientId := m.Payload.Marshall()

// 	Broker.topicList[topicID].consumerOffset[consumerId(clientId)] = 0
// 	return nil
// }

// func (b *BrokerServer) AddFollowerToTopic(m *Message, res *bool) error {
// 	topicID := m.Topic
// 	followerAddr := m.Payload.Marshall()

// 	tcpAddr, err := net.ResolveTCPAddr("tcp", string(followerAddr))
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, err.Error())
// 	}

// 	Broker.topicList[topicID].FollowerList = append(Broker.topicList[topicID].FollowerList, tcpAddr)
// 	return nil
// }

// func broadcastToFollowers(stub interface{}) error {
// 	return nil
// }
