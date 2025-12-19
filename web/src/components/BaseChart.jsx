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

export default function BaseChart({ title, processData, chartOptions = {} }) {
  const [chartData, setChartData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [syncing, setSyncing] = useState(false);

  const fetchAndProcessData = async () => {
    const response = await wanikaniAPI.getAssignmentSnapshots();
    const snapshots = response.data;
    const dates = Object.keys(snapshots).sort();
    return processData(snapshots, dates);
  };

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const data = await fetchAndProcessData();
        setChartData(data);
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
      const data = await fetchAndProcessData();
      setChartData(data);
    } catch (err) {
      console.error('Sync failed:', err);
      setError('Sync failed');
    } finally {
      setSyncing(false);
    }
  };

  const defaultOptions = {
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
    ...chartOptions
  };

  if (loading) {
    return (
      <div className="w-full h-full">
        <div className="bg-white rounded-2xl shadow-lg p-8 h-full flex flex-col">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-48 mb-8"></div>
            <div className="h-64 bg-gray-100 rounded flex-1"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full h-full">
        <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
          <p className="text-red-800">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      <div className="bg-white rounded-2xl shadow-lg h-full flex flex-col">
        <div className="flex justify-between items-center px-6 py-2">
          <h2 className="text-2xl font-bold text-gray-900">{title}</h2>
          <button
            onClick={handleSync}
            disabled={syncing}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
          >
            {syncing ? 'Syncing...' : 'Sync Data'}
          </button>
        </div>
        <div className="flex-1 pt-4 pl-4 bg-gray-100 rounded-lg mx-4 mb-4">
          <Bar data={chartData} options={defaultOptions} />
        </div>
      </div>
    </div>
  );
}