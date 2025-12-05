import axios from 'axios';

const API_BASE = import.meta.env.VITE_API_URL;
const API_TOKEN = import.meta.env.VITE_API_TOKEN;

if (!API_BASE) {
  throw new Error('VITE_API_URL environment variable is required');
}

if (!API_TOKEN) {
  throw new Error('VITE_API_TOKEN environment variable is required');
}

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Authorization': `Bearer ${API_TOKEN}`,
    'Content-Type': 'application/json',
  },
});

export const wanikaniAPI = {
  // Get assignment snapshots for charts
  getAssignmentSnapshots: (from, to) => {
    const params = {};
    if (from) params.from = from;
    if (to) params.to = to;
    const result = api.get('/assignments/snapshots', { params });
    //console.info('API getAssignmentSnapshots called with params:', params, 'Result:', result);
    return result;
  },

  // Get latest statistics
  getLatestStatistics: () => api.get('/statistics/latest'),

  // Get subjects
  getSubjects: (filters = {}) => api.get('/subjects', { params: filters }),

  // Get assignments
  getAssignments: (filters = {}) => api.get('/assignments', { params: filters }),

  // Trigger sync
  triggerSync: () => api.post('/sync'),

  // Health check
  getHealth: () => api.get('/health'),
};