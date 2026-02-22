package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func shutdownSignal() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch
		cancel()
	}()

	return ctx
}

func getWorkerCPU(pid int) (float64, error) {
	// Di Linux, file ini berisi statistik CPU proses tersebut
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}

	fmt.Printf(string(data), "log")

	// Parse data (sangat teknis, tapi library 'shirou/gopsutil' bisa bantu ini)
	// Untuk contoh ini, kita asumsikan pakai library biar gak muntah baca manual stat
	return float64(0), nil
}

func main() {
	// 1. Master bind ke Port 8080
	l, err := net.Listen("tcp", ":1111")
	if err != nil {
		log.Fatal(err)
	}

	shutdownCtx := shutdownSignal()

	// Ambil File Descriptor-nya
	f, _ := l.(*net.TCPListener).File()
	fd := int(f.Fd())
	log.Printf("MASTER nyala (PID: %d). Memegang FD: %d untuk Port :1111", os.Getpid(), fd)

	socketPath := "/tmp/http_mgr.sock"
	os.Remove(socketPath)
	unixLn, _ := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})

	log.Println("Menunggu Worker untuk mengambil kunci...")

	workerRegistry := make(map[string]*net.UnixConn)

	// 1. Buat loop abadi agar Master selalu siap nerima Worker baru
	go func() {
		for {
			conn, err := unixLn.AcceptUnix()
			if err != nil {
				log.Printf("Gagal Accept: %v", err)
				continue // Jika satu gagal, jangan stop Master-nya, lanjut nunggu yang lain
			}

			// 2. Kirim FD (Kunci)
			rights := syscall.UnixRights(fd)
			_, _, err = conn.WriteMsgUnix([]byte("PROCEED"), rights, nil)
			if err != nil {
				log.Printf("Gagal kirim FD: %v", err)
				conn.Close()
				continue
			}

			log.Println("Berhasil mengirim kunci ke satu Worker.")

			go func(c *net.UnixConn) {
				// Jangan lupa tutup koneksi kalau goroutine ini selesai
				defer c.Close()

				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					pid := strings.TrimSpace(scanner.Text())
					if pid == "" {
						continue
					}

					// Simpan ke Registry
					workerRegistry[pid] = c
					log.Printf("Worker Terdaftar | PID: %s | Total Worker: %d", pid, len(workerRegistry))
				}
			}(conn) // Kirim conn ke goroutine sebagai parameter
		}
	}()

	<-shutdownCtx.Done()

	fmt.Println("Cleansing")

	for k, v := range workerRegistry {
		v.Write([]byte("DIE"))
		log.Printf("PID %s send command DIE", k)

		if err := v.Close(); err != nil {
			fmt.Println(err)
		}

		log.Println("PID ", k, " is closed")
		delete(workerRegistry, k)
	}

	fmt.Println("Bye")
}
