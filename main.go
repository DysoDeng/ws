package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"ws/connection"
	"ws/message"
	"ws/server"
)

func main() {

	http.HandleFunc("/ws", server.WsHandler)
	http.HandleFunc("/ws/api", func(writer http.ResponseWriter, request *http.Request) {

		query := request.URL.Query()
		gid := query.Get("gid")

		if gid != "" {
			err := connection.WriteMessageAll(gid, message.OutMessage{
				Code: 0,
				Data: "这是接口发送的消息",
				Error: "",
				MessageType: message.TypeMessage,
			})
			log.Println(err)
		}

	})

	serve := http.Server{
		Addr: ":8888",
		Handler: nil,
	}

	go func() {
		// 服务连接
		if err := serve.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ws server listen: %s\n", err)
		}
	}()

	log.Println("ws server listen: 0.0.0.0:8888")

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("shutdown ws server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := serve.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown error:", err)
	}
	log.Println("server stop")
}