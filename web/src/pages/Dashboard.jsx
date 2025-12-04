import React, { useState, useEffect } from 'react';
import { wanikaniAPI } from '../services/api';

const Dashboard = () => {
  const [snapshots, setSnapshots] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await wanikaniAPI.getAssignmentSnapshots();
        setSnapshots(response.data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) return (
    <div className="flex items-center justify-center h-64">
      <div className="text-blue-600 text-lg">Loading...</div>
    </div>
  );
  
  if (error) return (
    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
      Error: {error}
    </div>
  );

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-8">Progress Overview</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {Object.entries(snapshots).map(([date, data]) => (
          <div key={date} className="bg-white rounded-lg shadow-md p-6">
            <h3 className="text-lg font-semibold text-gray-800 mb-4">{date}</h3>
            <div className="space-y-2">
              {Object.entries(data).map(([stage, counts]) => (
                <div key={stage} className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded">
                  <span className="font-medium capitalize text-gray-700">{stage}:</span>
                  <span className="text-blue-600 font-bold">{counts.total}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Dashboard;