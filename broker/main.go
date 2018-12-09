package main

import (
	"net"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	m"../lib/message"
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
)

type configSetting struct {
	BrokerNodeID      string
	BrokerIP          string
	ManagerIPs    []string
}

var config configSetting

var logger *govec.GoLog
var loggerOptions govec.GoLogOptions

/* readConfigJSON
 * Desc:
 *		read the configration from file into struct config
 *
 * @para configFile: relative url of file of configuration
 * @retrun: None
 */
func readConfigJSON(configFile string) error {
	configByte, err := ioutil.ReadFile(configFile)

	if err != nil {
		fmt.Println(err)
	}

	if err := json.Unmarshal(configByte, &config); err != nil {
		return err
	}

	return nil
}

// Initialize starts the node as a broker node in the network

func Initialize() error {
	configFilename := os.Args[1]

	if err := readConfigJSON(configFilename); err != nil {
		return err
	}

	logger = govec.InitGoVector(config.BrokerNodeID, fmt.Sprintf("%v-logfile", config.BrokerNodeID), govec.GetDefaultConfig())
	loggerOptions = govec.GetDefaultLogOptions()


	fmt.Println(config.BrokerIP)

	return nil
}

func registerBrokerWithManager() error{

	fmt.Println(config.ManagerIPs)

	managerAddr, err := net.ResolveTCPAddr("tcp", config.ManagerIPs[0])

	if err != nil {
		return err
	}

	rpcClient, err := vrpc.RPCDial("tcp", managerAddr.String(), logger, loggerOptions)
	defer rpcClient.Close()
	if err != nil {
		return err
	}

	message := m.Message{
		ID: config.BrokerNodeID,
		Text: config.BrokerIP,
	}

	var ack bool

	if err:=rpcClient.Call("ManagerRPCServer.RegisterBroker", message, &ack); err!=nil{
		return err
	}

	return nil
	
}


func main() {

	if len(os.Args) != 2 {
		fmt.Println("Please provide config filename. e.g. b1.json, b2.json")
		return
	}

	err := Initialize()
	checkError(err)

	err = InitBroker(config.BrokerIP)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
