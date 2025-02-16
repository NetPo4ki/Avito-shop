import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 },  // Ramp up
    { duration: '1m', target: 1000 },  // Test at 1k RPS
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99.99) < 50'], // 99.99% of requests should be under 50ms
    http_req_failed: ['rate<0.0001'],     // 99.99% success rate
  },
};

const BASE_URL = 'http://localhost:8080';
let token = '';

export function setup() {
  // Wait for services to be ready
  let retries = 5;
  while (retries > 0) {
    try {
      // Register and get token
      let res = http.post(`${BASE_URL}/api/auth`, JSON.stringify({
        username: 'loadtest',
        password: 'testpass'
      }));
      if (res.status === 200) {
        token = res.json('token');
        return { token };
      }
    } catch (e) {
      console.log(`Retrying setup... ${retries} attempts left`);
      retries--;
      sleep(5);
    }
  }
  throw new Error('Failed to setup load test after retries');
}

export default function(data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // Mix of different API calls
  let responses = http.batch([
    ['GET', `${BASE_URL}/api/info`, null, { headers }],
    ['GET', `${BASE_URL}/api/buy/test-item`, null, { headers }],
    ['POST', `${BASE_URL}/api/sendCoin`, 
      JSON.stringify({ toUser: 'recipient', amount: 100 }), { headers }],
  ]);

  check(responses[0], {
    'info status is 200': (r) => r.status === 200,
  });

  sleep(1);
} 