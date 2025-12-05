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

export default function DailyProportionsChart() {
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
        
        const apprenticeData = [];
        const guruData = [];
        const masterData = [];
        const enlightenedData = [];
        const burnedData = [];
        
        dates.forEach(date => {
          const dayData = snapshots[date];
          
          const apprenticeTotal = dayData.apprentice?.total || 0;
          const guruTotal = dayData.guru?.total || 0;
          const masterTotal = dayData.master?.total || 0;
          const enlightenedTotal = dayData.enlightened?.total || 0;
          const burnedTotal = dayData.burned?.total || 0;
          
          const grandTotal = apprenticeTotal + guruTotal + masterTotal + enlightenedTotal + burnedTotal;
          
          // Calculate percentages (avoid division by zero)
          if (grandTotal > 0) {
            apprenticeData.push((apprenticeTotal / grandTotal) * 100);
            guruData.push((guruTotal / grandTotal) * 100);
            masterData.push((masterTotal / grandTotal) * 100);
            enlightenedData.push((enlightenedTotal / grandTotal) * 100);
            burnedData.push((burnedTotal / grandTotal) * 100);
          } else {
            apprenticeData.push(0);
            guruData.push(0);
            masterData.push(0);
            enlightenedData.push(0);
            burnedData.push(0);
          }
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
        console.error(err);
        setError('Failed to load chart data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom',
      },
      title: {
        display: false,
      },
      tooltip: {
        callbacks: {
          label: function(context) {
            return `${context.dataset.label}: ${context.parsed.y.toFixed(1)}%`;
          }
        }
      }
    },
    scales: {
      x: {
        stacked: true,
      },
      y: {
        stacked: true,
        min: 0,
        max: 100,
        ticks: {
          callback: function(value) {
            return value + '%';
          }
        }
      },
    },
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
      <div className="bg-white rounded-2xl shadow-lg p-8 h-full flex flex-col">
        <div className="flex-1">
          <Bar data={chartData} options={options} />
        </div>
      </div>
    </div>
  );
}