// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data_conversion

import (
	pb "conversion/proto/msg"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"sync"
)

var (
	topic = "ws2ms_chat"
	addr  = "127.0.0.1:9092"
	//addr = "43.128.72.19:9092"
)

const (
	ZkAddr     = "127.0.0.1:2181"
	ZKSchema   = "openim"
	ZKUsername = ""
	ZKPassword = ""
	MsgName    = "Msg"
)

var consumer sarama.Consumer
var producerV2 sarama.SyncProducer
var wg sync.WaitGroup

//var msgRpcClient msg2.MsgClient

func init() {

	//Producer
	config := sarama.NewConfig()            // Instantiate a sarama Config
	config.Producer.Return.Successes = true // Whether to enable the successes channel to be notified after the message is sent successfully
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll        // Set producer Message Reply level 0 1 all
	config.Producer.Partitioner = sarama.NewHashPartitioner // Set the hash-key automatic hash partition. When sending a message, you must specify the key value of the message. If there is no key, the partition will be selected randomly

	client, err := sarama.NewSyncProducer([]string{addr}, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
	}
	producerV2 = client

	//Consumer
	consumerT, err := sarama.NewConsumer([]string{addr}, sarama.NewConfig())
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
	}
	consumer = consumerT

	//msgRpcClient = NewMessage()
}

func SendMessage() {
	// construct a message
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.StringEncoder("this is a test log")

	// Send a message
	pid, offset, err := producerV2.SendMessage(msg)
	if err != nil {
		fmt.Println("send msg failed, err:", err)
	}
	fmt.Printf("pid:%v offset:%v\n", pid, offset)
}

func GetMessage() {
	partitionList, err := consumer.Partitions(topic) // Get all partitions according to topic
	if err != nil {
		fmt.Printf("fail to get list of partition:err%v\n", err)
	}
	fmt.Println(partitionList)
	//var ch chan int

	if err != nil {
		fmt.Printf("rpc err:%s", err)
	}

	for partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetOldest)
		if err != nil {
			panic(err)
		}
		wg.Add(1)
		defer pc.AsyncClose()

		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			for msg := range pc.Messages() {
				Transfer([]*sarama.ConsumerMessage{msg})

				//V2
				//msgFromMQV2 := pb.MsgDataToMQ{}
				//err := proto.Unmarshal(msg.Value, &msgFromMQV2)
				//if err != nil {
				//	fmt.Printf("err:%s \n", err)
				//}
				//fmt.Printf("msg:%s \n", &msgFromMQV2)

				//V3
				//msgFromMQ := &sdkws.MsgData{}
				//err = proto.Unmarshal(msg.Value, msgFromMQ)
				//if err != nil {
				//	fmt.Printf("err:%s \n", err)
				//}
				//fmt.Printf("msg:%s \n", &msgFromMQ)
				fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			}

		}(pc)

	}
	wg.Wait()
	consumer.Close()
	//_ = <-ch
}

func Transfer(consumerMessages []*sarama.ConsumerMessage) {
	for i := 0; i < len(consumerMessages); i++ {
		msgFromMQV2 := pb.MsgDataToMQ{}
		err := proto.Unmarshal(consumerMessages[i].Value, &msgFromMQV2)
		if err != nil {
			fmt.Printf("err:%s \n", err)
		}
		fmt.Printf("msg:%s \n", &msgFromMQV2)
		//fmt.Printf("rpcClient:%s \n", msgRpcClient)
		//msgRpcClient.SendMsg(context.Background(),msgFromMQV2)
	}
}

// GetMsgRpcService Convenient for detachment
//func GetMsgRpcService() (rpcclient.MessageRpcClient, error) {
//	client, err := openKeeper.NewClient([]string{ZkAddr}, ZKSchema,
//		openKeeper.WithFreq(time.Hour), openKeeper.WithRoundRobin(), openKeeper.WithUserNameAndPassword(ZKUsername,
//			ZKPassword), openKeeper.WithTimeout(10))
//	msgClient := rpcclient.NewMessageRpcClient(client)
//	if err != nil {
//		return msgClient, err
//	}
//	return msgClient, nil
//}

//func NewMessage() msg2.MsgClient {
//	discov, err := openkeeper.NewClient([]string{ZkAddr}, ZKSchema,
//		openkeeper.WithFreq(time.Hour), openkeeper.WithRoundRobin(), openkeeper.WithUserNameAndPassword(ZKUsername,
//			ZKPassword), openkeeper.WithTimeout(10))
//	conn, err := discov.GetConn(context.Background(), MsgName)
//	if err != nil {
//		panic(err)
//	}
//	client := msg2.NewMsgClient(conn)
//	return client
//}
