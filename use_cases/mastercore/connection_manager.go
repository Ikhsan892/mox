package mastercore

import "net"

type ConnectionManager interface {
	// OpenListener membuka port TCP baru (misal: ":80") dan menyimpannya di registry
	OpenListener(name string, address string) (net.Listener, error)

	// CloseListener menutup port tertentu dan menghapusnya dari registry
	CloseListener(name string) error

	// GetListener mengambil listener yang sudah ada berdasarkan nama (e.g., "web-http")
	GetListener(name string) (net.Listener, bool)

	// --- FD Operations ---

	// GetRawFD mengambil angka File Descriptor dari listener tertentu
	// untuk siap dikirim via Unix Domain Socket (SCM_RIGHTS)
	GetRawFD(name string) (uintptr, error)

	// --- Inventory & Health ---

	// ActivePorts mengembalikan daftar nama port yang sedang dikelola Master
	ActivePorts() []string
}
