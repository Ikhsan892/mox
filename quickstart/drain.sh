# 1. Masuk ke container haproxy-1
#    (Nama container sesuaikan kalau beda, cek 'docker ps')
docker exec -it quickstart-haproxy-1-1 bash

# 2. Di dalam container: Install socat (karena image haproxy minimalis)
apt-get update && apt-get install -y socat

# 3. Drain Traffic (Set MaxConn ke 0)
#    Ini simulasi "graceful shutdown" -> stop accept new connection
echo "set maxconn frontend main 0" | socat stdio /var/lib/haproxy/sockets/stats-haproxy-1.sock

# 4. Verifikasi
#    Cek traffic di window PowerShell sebelah, harusnya pindah semua ke haproxy-2.
#    Cek status di stats socket:
echo "show stat" | socat stdio /var/lib/haproxy/sockets/stats-haproxy-1.sock | grep main
