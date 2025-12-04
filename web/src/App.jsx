import React from 'react';
import Dashboard from './pages/Dashboard.jsx';
import ItemSpreadCard from './components/ItemSpreadCard.jsx';

const todayData = {
  apprentice: { radical: 4, kanji: 7, vocabulary: 17, total: 28 },
  guru: { radical: 12, kanji: 31, vocabulary: 105, total: 148 },
  master: { radical: 30, kanji: 35, vocabulary: 87, total: 152 },
  enlightened: { radical: 58, kanji: 72, vocabulary: 174, total: 304 },
  burned: { radical: 55, kanji: 32, vocabulary: 43, total: 130 },
};

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
        <div className="min-h-screen bg-gray-50 py-12">
          <ItemSpreadCard data={todayData} />
        </div>
      </main>
    </div>
  );
}

export default App;