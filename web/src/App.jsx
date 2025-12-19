import { BrowserRouter, Routes, Route } from 'react-router-dom';
import DashboardPage from './pages/DashboardPage.jsx';
import AboutPage from './pages/AboutPage.jsx';
import Navigation from './components/Navigation.jsx';

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-blue-600 text-white shadow-lg">
          <div className="mx-auto px-4 py-6">
            <h1 className="text-3xl font-bold">WaniKani Dashboard</h1>
          </div>
        </header>
        
        <Navigation />
        
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/about" element={<AboutPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;