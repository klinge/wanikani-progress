import React, { useState, useEffect } from 'react';
import { Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { wanikaniAPI } from '../services/api';

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

export default function DailyTotalsChart() {
  const [chartData, setChartData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [syncing, setSyncing] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await wanikaniAPI.getAssignmentSnapshots();
        
        const snapshots = response.data;
        const dates = Object.keys(snapshots).sort();
        
        //console.log('Available dates:', dates);
        //console.log('Snapshots data:', snapshots);
        
        const apprenticeData = [];
        const guruData = [];
        const masterData = [];
        const enlightenedData = [];
        const burnedData = [];
        
        dates.forEach(date => {
          const dayData = snapshots[date];
          
          apprenticeData.push(dayData.apprentice?.total || 0);
          guruData.push(dayData.guru?.total || 0);
          masterData.push(dayData.master?.total || 0);
          enlightenedData.push(dayData.enlightened?.total || 0);
          burnedData.push(dayData.burned?.total || 0);
        });

        setChartData({
          labels: dates,
          datasets: [
            {
              label: 'Apprentice',
              data: apprenticeData,
              backgroundColor: '#00AAFF',
            },
            {
              label: 'Guru',
              data: guruData,
              backgroundColor: '#FF00AA',
            },
            {
              label: 'Master',
              data: masterData,
              backgroundColor: '#294ddb',
            },
            {
              label: 'Enlightened',
              data: enlightenedData,
              backgroundColor: '#0093dd',
            },
            {
              label: 'Burned',
              data: burnedData,
              backgroundColor: '#434343',
            },
          ],
        });
      } catch (err) {
        console.error(err);
        setError('Failed to load chart data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleSync = async () => {
    try {
      setSyncing(true);
      await wanikaniAPI.triggerSync();
      // Refetch data after sync
      const response = await wanikaniAPI.getAssignmentSnapshots();
      const snapshots = response.data;
      const dates = Object.keys(snapshots).sort();
      
      const apprenticeData = [];
      const guruData = [];
      const masterData = [];
      const enlightenedData = [];
      const burnedData = [];
      
      dates.forEach(date => {
        const dayData = snapshots[date];
        
        apprenticeData.push(dayData.apprentice?.total || 0);
        guruData.push(dayData.guru?.total || 0);
        masterData.push(dayData.master?.total || 0);
        enlightenedData.push(dayData.enlightened?.total || 0);
        burnedData.push(dayData.burned?.total || 0);
      });

      setChartData({
        labels: dates,
        datasets: [
          {
            label: 'Apprentice',
            data: apprenticeData,
            backgroundColor: '#dd1166',
          },
          {
            label: 'Guru',
            data: guruData,
            backgroundColor: '#882d9e',
          },
          {
            label: 'Master',
            data: masterData,
            backgroundColor: '#294ddb',
          },
          {
            label: 'Enlightened',
            data: enlightenedData,
            backgroundColor: '#0093dd',
          },
          {
            label: 'Burned',
            data: burnedData,
            backgroundColor: '#434343',
          },
        ],
      });
    } catch (err) {
      console.error('Sync failed:', err);
      setError('Sync failed');
    } finally {
      setSyncing(false);
    }
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    layout: {
      padding: 0
    },
    plugins: {
      legend: {
        position: 'bottom',
      },
      title: {
        display: false,
        text: 'Daily Item Totals',
      },
    },
    scales: {
      x: {
        stacked: true,
        grid: {
          color: '#f3f4f6'
        }
      },
      y: {
        stacked: true,
        grid: {
          color: '#f3f4f6'
        }
      },
    },
  };

  if (loading) {
    return (
      <div className="w-full max-w-4xl mx-auto p-8">
        <div className="bg-white rounded-2xl shadow-lg p-8">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-48 mb-8"></div>
            <div className="h-64 bg-gray-100 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full max-w-4xl mx-auto p-8">
        <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
          <p className="text-red-800">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      <div className="bg-white rounded-2xl shadow-lg h-full flex flex-col">
        <div className="flex justify-between items-center px-6 pt-6 pb-4">
          <h2 className="text-2xl font-bold text-gray-900">Daily Item Totals</h2>
          <button
            onClick={handleSync}
            disabled={syncing}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
          >
            {syncing ? 'Syncing...' : 'Sync Data'}
          </button>
        </div>
        <div className="flex-1 pt-4 pl-4 bg-gray-100 rounded-lg mx-4 mb-4">
          <Bar data={chartData} options={options} />
        </div>
      </div>
    </div>
  );
}