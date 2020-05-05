package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func consumeMsg() {
	// make a new reader that consumes from topic-A, partition 0, at offset 42
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   kafkaURLs,
		Topic:     topic,
		Partition: 0,
		//MinBytes:  10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	// 设置偏移量,从哪个位置开始消费
	r.SetOffset(42)
	defer r.Close()

	log.Println(1111)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("error: ", err)
			break
		}

		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}
}

/**
kafka-go also supports Kafka consumer groups including broker managed offsets.
To enable consumer groups, simply specify the GroupID in the ReaderConfig.

ReadMessage automatically commits offsets when using consumer groups.
// 会自动提交偏移量
*/
func consumeGroupMsg(groupId string) {
	// make a new reader that consumes from topic-A, partition 0, at offset 42
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   kafkaURLs,
		Topic:     topic,
		Partition: 0,
		GroupID:   groupId,
		//MinBytes:  10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxAttempts: 3,    // 最大重试次数
		MaxWait:     10 * time.Second,
	})

	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("error: ", err)
			break
		}

		fmt.Println("current group_id: ", groupId)
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}
}

/**
需要监听kafka ip:port
修改config/server.properties配置文件，更改如下
把31行的注释去掉，listeners=PLAINTEXT://:9092
把36行的注释去掉，把advertised.listeners值改为PLAINTEXT://host_ip:9092
然后重新启动kafka
*/
var (
	kafkaURLs = []string{
		"192.168.0.11:9092",
	}

	topic = "test"
)

func main() {
	log.Println("===start consume msg=====")
	//go consumeMsg()

	// 指定多个消费组模式，消费topic中的消息
	go consumeGroupMsg("mytest-1")
	go consumeGroupMsg("mytest-2")

	ch := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// recivie signal to exit main goroutine
	//window signal
	// signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGHUP)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2, os.Interrupt, syscall.SIGHUP)

	// Block until we receive our signal.
	<-ch

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// if your application should wait for other services
	// to finalize based on context cancellation.
	<-ctx.Done()

	log.Println("shutting down")
}

/**
消费者：
消费者使用一个 消费组 名称来进行标识，发布到topic中的每条记录被分配给订阅消费组中的一个消费者实例.消费者实例可以分布在多个进程中或者多个机器上。
如果所有的消费者实例在同一消费组中，消息记录会负载平衡到每一个消费者实例.
如果所有的消费者实例在不同的消费组中，每条消息记录会广播到所有的消费者进程.
*/
/**
2020/05/04 23:09:44 ===start consume msg=====
current group_id:  mytest-2
message at offset 73: mytest:hello = hello,eee123
current group_id:  mytest-2
message at offset 74: mytest:hello = hello,eee123
current group_id:  mytest-1
message at offset 73: mytest:hello = hello,eee123
current group_id:  mytest-1
message at offset 74: mytest:hello = hello,eee123
current group_id:  mytest-2
message at offset 75: mytest:hello = hello,eee1234444
current group_id:  mytest-1
message at offset 75: mytest:hello = hello,eee1234444
*/
