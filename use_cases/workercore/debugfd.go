package workercore

import (
	"fmt"
	"syscall"
)

// Panggil ini: InspectFD(int(w.ExtraFile.Fd()), "Socket Warisan")
func InspectFD(fd int, label string) {
	fmt.Printf("\n--- 🕵️‍♂️ INSPEKSI FD: %s (No: %d) ---\n", label, fd)

	// 1. Cek Validitas & Tipe (Stat)
	var stat syscall.Stat_t
	err := syscall.Fstat(fd, &stat)
	if err != nil {
		fmt.Printf("❌ ERROR: FD %d KOSONG/TUTUP/RUSAK! (%v)\n", fd, err)
		return
	}

	// 2. Cek Mode (Apakah Socket, File, atau Pipe?)
	mode := stat.Mode
	isSocket := (mode & syscall.S_IFMT) == syscall.S_IFSOCK
	isFile := (mode & syscall.S_IFMT) == syscall.S_IFREG
	isPipe := (mode & syscall.S_IFMT) == syscall.S_IFIFO

	fmt.Printf("   ├─ Status: VALID (Open)\n")
	fmt.Printf("   ├─ Mode Raw: %o\n", mode)

	if isSocket {
		fmt.Printf("   ✅ Tipe: SOCKET (Aman!)\n")

		// 3. Cek Detail Socket (IP & Port)
		// Ini bakal gagal kalau socketnya bukan Network Socket (misal Unix Socket)
		sa, err := syscall.Getsockname(fd)
		if err == nil {
			// Coba cast ke IPv4
			if v4, ok := sa.(*syscall.SockaddrInet4); ok {
				fmt.Printf("   ├─ Bind Address: %d.%d.%d.%d:%d\n",
					v4.Addr[0], v4.Addr[1], v4.Addr[2], v4.Addr[3], v4.Port)
			} else if v6, ok := sa.(*syscall.SockaddrInet6); ok {
				fmt.Printf("   ├─ Bind Address: [IPv6]:%d\n", v6.Port)
			} else {
				fmt.Printf("   ├─ Bind Address: (Unix/Other)\n")
			}
		} else {
			fmt.Printf("   ⚠️  Gagal ambil Sockname: %v\n", err)
		}

	} else if isFile {
		fmt.Printf("   ❌ Tipe: REGULAR FILE (Salah barang!)\n")
	} else if isPipe {
		fmt.Printf("   ❌ Tipe: PIPE (Salah barang!)\n")
	} else {
		fmt.Printf("   ❓ Tipe: UNKNOWN\n")
	}

	fmt.Println("----------------------------------------\n")
}
