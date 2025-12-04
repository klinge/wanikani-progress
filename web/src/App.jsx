import React from 'react';
import Dashboard from './pages/Dashboard.jsx';

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-blue-600 text-white shadow-lg">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <h1 className="text-3xl font-bold">WaniKani Dashboard</h1>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 py-8">
        <Dashboard />
      </main>
    </div>
  );
}

export default App;