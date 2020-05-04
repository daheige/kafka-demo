package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	kafka "github.com/segmentio/kafka-go"
)

// producerHandler 处理kafka消息发送
func producerHandler(ctx context.Context, kafkaWriter *kafka.Writer, key string, body []byte) error {
	// kafka message
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("mytest:%s", key)),
		Value: body,
	}

	err := kafkaWriter.WriteMessages(ctx, msg)

	if err != nil {
		return err
	}

	log.Println("send msg success")
	return nil
}

// getKafkaWriter 获取kafka writer fd
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

// sendMsg 发送消息http handlerFunc
// http://localhost:1337/send-msg?name=daheige
func sendMsg(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		w.Write([]byte("name param is empty"))
		return
	}

	kafkaWriter := getKafkaWriter(kafkaURLs, topic)
	defer kafkaWriter.Close()

	body := []byte("hello," + name)
	err := producerHandler(r.Context(), kafkaWriter, "hello", body)
	if err != nil {
		w.Write([]byte("send msg error:" + err.Error()))
		return
	}

	w.Write([]byte("send msg success"))
}

// postMsg post body 消息接收
// curl localhost:1337/post-msg -X POST -d '{"name":"daheige"}' --header "Content-Type: application/json"
func postMsg(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("read body error: " + err.Error()))
		return
	}

	defer r.Body.Close()

	kafkaWriter := getKafkaWriter(kafkaURLs, topic)
	defer kafkaWriter.Close()

	err = producerHandler(r.Context(), kafkaWriter, "hello", body)
	if err != nil {
		w.Write([]byte("send msg error:" + err.Error()))
		return
	}

	w.Write([]byte("send msg success"))

}

var kafkaURLs = []string{
	"192.168.0.11:9092",
}

var topic = "test"

func main() {
	appPort := 1337
	router := mux.NewRouter()

	//与http.ServerMux不同的是mux.Router是完全的正则匹配,设置路由路径/index/，如果访问路径/idenx/hello会返回404
	//设置路由路径为/index/访问路径/index也是会报404的,需要设置r.StrictSlash(true), /index/与/index才能匹配
	router.StrictSlash(true)

	//健康检查
	router.HandleFunc("/check", HealthCheck)

	// 测试kafka发送消息
	router.HandleFunc("/send-msg", sendMsg).Methods("GET")

	// 接受post body message
	router.HandleFunc("/post-msg", postMsg).Methods("POST")

	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", appPort),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	//在独立携程中运行
	log.Println("server run on: ", appPort)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("server listen error:", err)
		}
	}()

	//mux平滑重启
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
	go server.Shutdown(ctx) //在独立的携程中关闭服务器
	<-ctx.Done()

	log.Println("shutting down")
}

// HealthCheck A very simple health check.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	w.Write([]byte(`{"alive": true}`))
}

/**
运行后，就可以发消息到kafka的test topic上
打开消费者的shell
[zhuwei@daheige kafka_2.11-1.0.0]$ bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic test
hello,daheige
hello,daheige
*/
