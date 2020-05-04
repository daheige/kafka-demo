package main

import (
	"context"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func producerHandler(kafkaWriter *kafka.Writer) {
	body := "hello,daheige"
	name := "hello"
	// kafka message
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("mytest:%s", name)),
		Value: []byte(body),
	}

	ctx := context.Background()
	err := kafkaWriter.WriteMessages(ctx, msg)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("send msg success")
}

func getKafkaWriter(kafkaURLs []string, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:         kafkaURLs,
		Topic:           topic,
		Balancer:        &kafka.LeastBytes{},
		MaxAttempts:     3,               // 最大重试次数
		WriteTimeout:    3 * time.Second, // 写入超时
		IdleConnTimeout: 6 * time.Minute, // 默认空闲时间
	})
}

/**
需要监听kafka ip:port
修改config/server.properties配置文件，更改如下
把31行的注释去掉，listeners=PLAINTEXT://:9092
把36行的注释去掉，把advertised.listeners值改为PLAINTEXT://host_ip:9092
然后重新启动kafka
*/
func main() {
	// get kafka writer using environment variables.
	kafkaURLs := []string{
		"192.168.0.11:9092",
	}

	topic := "test"
	kafkaWriter := getKafkaWriter(kafkaURLs, topic)

	defer kafkaWriter.Close()

	producerHandler(kafkaWriter)
}

/**
运行后，就可以发消息到kafka的test topic上
打开消费者的shell
[zhuwei@daheige kafka_2.11-1.0.0]$ bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic test
hello,daheige
hello,daheige
*/
