import ItemSpreadCard from '../components/ItemSpreadCard.jsx';
import DailyTotalsChart from '../components/DailyTotalsChart.jsx';
import DailyProportionsChart from '../components/DailyProportionsChart.jsx';

function DashboardPage() {
  return (
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
  );
}

export default DashboardPage;