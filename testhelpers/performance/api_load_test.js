import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const apiResponseTime = new Trend('api_response_time');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users over 2 minutes
    { duration: '5m', target: 100 },   // Stay at 100 users for 5 minutes
    { duration: '2m', target: 200 },   // Ramp up to 200 users over 2 minutes
    { duration: '5m', target: 200 },   // Stay at 200 users for 5 minutes
    { duration: '2m', target: 500 },   // Ramp up to 500 users over 2 minutes
    { duration: '5m', target: 500 },   // Stay at 500 users for 5 minutes
    { duration: '2m', target: 0 },     // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate should be below 10%
    errors: ['rate<0.1'],             // Custom error rate metric
  },
  setupTimeout: '30s',
  teardownTimeout: '30s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Setup function to initialize test data
export function setup() {
  // Create a test user and get authentication token
  const loginPayload = {
    email: 'test@example.com',
    password: 'test_password',
  };

  const loginResponse = http.post(`${BASE_URL}/v1/auth/login`, JSON.stringify(loginPayload), {
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (loginResponse.status !== 200) {
    console.warn('Login failed, proceeding with public endpoints only');
    return { token: null };
  }

  const responseBody = JSON.parse(loginResponse.body);
  return { token: responseBody.token };
}

// Default function (main test scenario)
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
  };

  if (data.token) {
    headers['Authorization'] = `Bearer ${data.token}`;
  }

  // Health check endpoint (lightweight)
  const healthResponse = http.get(`${BASE_URL}/health`, { headers });
  check(healthResponse, {
    'health status is 200': (r) => r.status === 200,
    'health response time < 200ms': (r) => r.timings.duration < 200,
  });
  errorRate.add(healthResponse.status !== 200);

  // List products (read-heavy operation)
  const productsResponse = http.get(`${BASE_URL}/v1/products?limit=20&offset=0`, { headers });
  check(productsResponse, {
    'products status is 200': (r) => r.status === 200,
    'products response time < 300ms': (r) => r.timings.duration < 300,
  });
  errorRate.add(productsResponse.status !== 200);
  apiResponseTime.add(productsResponse.timings.duration);

  // Search products (complex query)
  const searchResponse = http.get(`${BASE_URL}/v1/products/search?q=fertilizer&limit=10`, { headers });
  check(searchResponse, {
    'search status is 200': (r) => r.status === 200,
    'search response time < 500ms': (r) => r.timings.duration < 500,
  });
  errorRate.add(searchResponse.status !== 200);
  apiResponseTime.add(searchResponse.timings.duration);

  // Get product analytics (aggregation)
  const analyticsResponse = http.get(`${BASE_URL}/v1/products/analytics`, { headers });
  check(analyticsResponse, {
    'analytics status is 200': (r) => r.status === 200,
    'analytics response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  errorRate.add(analyticsResponse.status !== 200);
  apiResponseTime.add(analyticsResponse.timings.duration);

  // Simulate periodic busy operations (10% of requests)
  if (Math.random() < 0.1) {
    const heavyResponse = http.get(`${BASE_URL}/v1/products/search?q=*&limit=50&offset=0`, { headers });
    check(heavyResponse, {
      'heavy search status is 200': (r) => r.status === 200,
      'heavy search response time < 2000ms': (r) => r.timings.duration < 2000,
    });
    errorRate.add(heavyResponse.status !== 200);
  }

  // Simulate write operations (5% of requests require authenticated user)
  if (Math.random() < 0.05 && data.token) {
    const createProductResponse = http.post(
      `${BASE_URL}/v1/products`,
      JSON.stringify({
        name: `Test Product ${Math.random()}`,
        quantity: Math.floor(Math.random() * 1000),
        unit_price: parseFloat((Math.random() * 100).toFixed(2)),
        description: 'Performance test product',
      }),
      { headers }
    );
    check(createProductResponse, {
      'create product status is 201': (r) => r.status === 201,
      'create product response time < 1000ms': (r) => r.timings.duration < 1000,
    });
    errorRate.add(createProductResponse.status !== 201);
  }

  // Random sleep to simulate user behavior
  sleep(Math.random() * 2 + 1); // Sleep between 1-3 seconds
}

// Teardown function for cleanup
export function teardown(data) {
  console.log('Performance test completed');
  console.log(`Test processed by VUs: ${__ENV.VUs || 'unknown'}`);
  if (data.token) {
    // Optional: Revoke test tokens or cleanup test data
  }
}