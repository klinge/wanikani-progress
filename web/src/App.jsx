import React from 'react';
import Dashboard from './pages/Dashboard.jsx';
import ItemSpreadCard from './components/ItemSpreadCard.jsx';
import DailyTotalsChart from './components/DailyTotalsChart.jsx';

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-blue-600 text-white shadow-lg">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <h1 className="text-3xl font-bold">WaniKani Dashboard</h1>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="grid lg:grid-cols-2 gap-8 max-w-7xl mx-auto py-8">
          {/* Left column – Item Spread */}
          <div className="w-full h-96">
            <ItemSpreadCard />
          </div>

          {/* Right column – stacked bar diagram with item count per srs_level */}
          <div className="w-full h-96">
            <DailyTotalsChart />
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;