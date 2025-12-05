import React from 'react';
import Dashboard from './pages/Dashboard.jsx';
import ItemSpreadCard from './components/ItemSpreadCard.jsx';
import DailyTotalsChart from './components/DailyTotalsChart.jsx';
import DailyProportionsChart from './components/DailyProportionsChart.jsx';

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-blue-600 text-white shadow-lg">
        <div className="mx-auto px-4 py-6">
          <h1 className="text-3xl font-bold">WaniKani Dashboard</h1>
        </div>
      </header>
      <main className="mx-auto px-2 sm:px-4 py-8">
        <div className="grid lg:grid-cols-2 gap-4 lg:gap-8 mx-auto py-8">
          {/* Left column – Item Spread */}
          <div className="w-full h-80 lg:h-96">
            <ItemSpreadCard />
          </div>

          {/* Right column – stacked bar diagram with item count per srs_level */}
          <div className="w-full h-80 lg:h-96">
            <DailyTotalsChart />
          </div>

          {/* Third component - spans both columns */}
          <div className="w-full h-80 lg:h-120 lg:col-span-2">
            <DailyProportionsChart />
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;