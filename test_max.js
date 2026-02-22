import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate } from 'k6/metrics';

// Custom Metrics buat ngitung traffic per Worker
let requestsToWorkerA = new Counter('req_worker_A'); // Misal PID 1001
let requestsToWorkerB = new Counter('req_worker_B'); // Misal PID 1002
let errorRate = new Rate('errors');

export const options = {
  // Skenario Load Test
  stages: [
    { duration: '30s', target: 50 }, // Ramp up ke 50 user
    { duration: '1m', target: 50 },  // Tahan di 50 user (SAAT INI LU LAKUKAN SIMULASI MAXCONN)
    { duration: '10s', target: 0 },  // Ramp down
  ],
  thresholds: {
    errors: ['rate<0.10'], // Gagal kalau error rate > 1%
  },
};

export default function() {
  const url = 'http://127.0.0.1:1111'; // URL Mox Frontend

  const res = http.get(url);

  // 1. Cek Connection Loss / Error
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
  });

  errorRate.add(!success);

  sleep(0.1); // Jeda dikit biar gak flood parah
}
