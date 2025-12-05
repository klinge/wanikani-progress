import React from 'react';
import BaseChart from './BaseChart';

export default function DailyTotalsChart() {
  const processData = (snapshots, dates) => {
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

  return (
    <BaseChart 
      title="Daily Item Totals" 
      processData={processData} 
    />
  );
}