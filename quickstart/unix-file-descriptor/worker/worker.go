package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func shutdownSignal() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch
		cancel()
	}()

	return ctx, cancel
}

func startHTTPServer(l net.Listener, pid int) *http.Server {
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			msg := fmt.Sprintf("Request ditangani oleh PID: %d\n", pid)
			log.Println(msg)
			w.Write([]byte(msg))
		}),
	}

	go func() {
		fmt.Printf("WORKER %d: HTTP Server mulai melayani...\n", pid)
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Listen error: %v\n", err)
		}
	}()

	return srv
}

func listenMaster(conn *net.UnixConn, pid int, cancelFunc context.CancelFunc) {
	report := fmt.Sprintf("%d\n", pid)

	conn.Write([]byte(report))
	fmt.Printf("Worker %d: Berhasil lapor ke Master\n", pid)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())

		fmt.Printf("[WORKER %d] Nerima Instruksi: %s\n", pid, command)

		if command == "DIE" {
			cancelFunc()

			fmt.Printf("[WORKER %d] Nerima Instruksi: %s buat matiin server euy \n", pid, command)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Master hilang/mati. Worker ikut pamit!", err.Error())
		os.Exit(0) // Worker mematikan diri sendiri
	}
}

func runCMD() {
}

func main() {
	pid := os.Getpid()

	signalCTX, cancel := shutdownSignal()

	// 1. Ambil FD dari Master lewat Unix Socket
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: "/tmp/http_mgr.sock", Net: "unix"})
	if err != nil {
		fmt.Printf("Worker %d gagal konek ke Master: %v\n", pid, err)
		return
	}

	oob := make([]byte, syscall.CmsgSpace(4))
	dummy := make([]byte, 7)

	_, oobn, _, _, _ := conn.ReadMsgUnix(dummy, oob)

	// 2. Bongkar FD-nya
	msgs, _ := syscall.ParseSocketControlMessage(oob[:oobn])

	fmt.Println("dapet message dari parent ", string(dummy))

	go listenMaster(conn, pid, cancel)

	fds, _ := syscall.ParseUnixRights(&msgs[0])

	for i, m := range msgs {
		fmt.Printf("Pesan ke-%d:\n", i)
		fmt.Printf("  Level: %d, Type: %d\n", m.Header.Level, m.Header.Type)
		fmt.Printf("  Raw Data (Bytes): %v\n", m.Data)
	}

	if len(fds) > 0 {
		fmt.Printf("FD yang berhasil diekstrak: %d\n", fds[0])
	}

	receivedFd := fds[0]

	// 3. Ubah FD mentah jadi Listener
	file := os.NewFile(uintptr(receivedFd), "net-socket")
	l, _ := net.FileListener(file)

	srv := startHTTPServer(l, pid)

	<-signalCTX.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Srv Shutdown Error: %v\n", err)
	}

	fmt.Println("[WORKER] Bye bro, semua koneksi sudah beres.")
}
