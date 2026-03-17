import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pdfProcessingTime = new Rate('pdf_processing_time');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 },   // Ramp up to 10 users
    { duration: '5m', target: 10 },   // Stay at 10 users
    { duration: '2m', target: 50 },   // Ramp up to 50 users
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],             // Custom error rate must be below 10%
  },
};

// Test data
const testPDFs = [
  'test-data/real-automotive-report.pdf',
  'test-data/test.pdf',
];

const searchQueries = [
  'engine performance metrics',
  'fuel consumption ratings',
  'safety ratings NCAP',
  'Range Rover Sport specifications',
  '0-60 mph acceleration',
  'top speed performance',
  'engine displacement',
  'power output torque',
  'city driving mpg',
  'highway driving efficiency',
];

// Main test function
export default function() {
  const baseURL = __ENV.BASE_URL || 'http://localhost:8080';
  
  // Random user behavior
  const userType = Math.random();
  
  if (userType < 0.3) {
    // 30% - Document upload
    testDocumentUpload(baseURL);
  } else if (userType < 0.7) {
    // 40% - Semantic search
    testSemanticSearch(baseURL);
  } else {
    // 30% - RAG Q&A
    testRAGQA(baseURL);
  }
  
  sleep(Math.random() * 3 + 1); // Random sleep between 1-4 seconds
}

// Test document upload functionality
function testDocumentUpload(baseURL) {
  const startTime = Date.now();
  
  // Simulate file upload (in real test, you'd need actual file data)
  const formData = {
    document: 'test-pdf-content',
    type: 'automotive_test',
  };
  
  const response = http.post(`${baseURL}/v1/documents/upload`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  
  const processingTime = Date.now() - startTime;
  
  const success = check(response, {
    'upload status is 202': (r) => r.status === 202,
    'response has document_id': (r) => r.json('document_id') !== undefined,
    'response has s3_key': (r) => r.json('s3_key') !== undefined,
  });
  
  if (!success) {
    errorRate.add(1);
  }
  
  pdfProcessingTime.add(processingTime);
}

// Test semantic search functionality
function testSemanticSearch(baseURL) {
  const query = searchQueries[Math.floor(Math.random() * searchQueries.length)];
  const searchType = Math.random() < 0.5 ? 'cell' : 'figure';
  const topK = Math.floor(Math.random() * 15) + 5; // 5-20 results
  
  const payload = JSON.stringify({
    query: query,
    top_k: topK,
    type: searchType,
  });
  
  const response = http.post(`${baseURL}/v1/search/semantic`, payload, {
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  const success = check(response, {
    'search status is 200': (r) => r.status === 200,
    'response has results': (r) => r.json('results') !== undefined,
    'response has total': (r) => r.json('total') !== undefined,
    'results array exists': (r) => Array.isArray(r.json('results')),
  });
  
  if (!success) {
    errorRate.add(1);
  }
}

// Test RAG Q&A functionality
function testRAGQA(baseURL) {
  const question = searchQueries[Math.floor(Math.random() * searchQueries.length)];
  const topK = Math.floor(Math.random() * 10) + 3; // 3-13 results
  
  const payload = JSON.stringify({
    question: `What are the ${question} for the Range Rover Sport?`,
    top_k: topK,
  });
  
  const response = http.post(`${baseURL}/v1/ask`, payload, {
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  const success = check(response, {
    'ask status is 200': (r) => r.status === 200,
    'response has answer': (r) => r.json('answer') !== undefined,
    'response has citations': (r) => r.json('citations') !== undefined,
    'response has model': (r) => r.json('model') !== undefined,
    'citations array exists': (r) => Array.isArray(r.json('citations')),
  });
  
  if (!success) {
    errorRate.add(1);
  }
}

// Test health endpoint
export function setup() {
  const baseURL = __ENV.BASE_URL || 'http://localhost:8080';
  
  // Check if system is healthy before starting tests
  const healthResponse = http.get(`${baseURL}/v1/health`);
  
  check(healthResponse, {
    'system is healthy': (r) => r.status === 200 && r.json('status') === 'ok',
  });
  
  if (healthResponse.status !== 200) {
    throw new Error('System is not healthy, aborting load test');
  }
  
  console.log('System is healthy, starting load test...');
}

// Teardown function
export function teardown(data) {
  console.log('Load test completed');
  console.log('Final error rate:', errorRate.value);
  console.log('Average PDF processing time:', pdfProcessingTime.value);
}
