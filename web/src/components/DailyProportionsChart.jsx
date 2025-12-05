import React from 'react';
import BaseChart from './BaseChart';

export default function DailyProportionsChart() {
  const processData = (snapshots, dates) => {
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

    return {
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
    };
  };

  const chartOptions = {
    plugins: {
      tooltip: {
        callbacks: {
          label: function(context) {
            return `${context.dataset.label}: ${context.parsed.y.toFixed(1)}%`;
          }
        }
      }
    },
    scales: {
      y: {
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

  return (
    <BaseChart 
      title="SRS Stage Proportions" 
      processData={processData}
      chartOptions={chartOptions}
    />
  );
}