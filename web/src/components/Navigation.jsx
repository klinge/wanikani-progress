import { NavLink } from 'react-router-dom';
import { BarChart3, Info } from 'lucide-react';

function Navigation() {
  const navigationItems = [
    { path: '/', label: 'Dashboard', icon: BarChart3 },
    { path: '/about', label: 'About', icon: Info }
  ];

  return (
    <nav className="bg-white shadow-md">
      <div className="mx-auto px-4">
        <div className="flex space-x-1 sm:space-x-4">
          {navigationItems.map(({ path, label, icon: Icon }) => (
            <NavLink
              key={path}
              to={path}
              className={({ isActive }) =>
                `flex items-center gap-2 px-3 py-4 text-sm sm:text-base font-medium transition-colors duration-200 border-b-2 ${
                  isActive
                    ? 'text-blue-600 border-blue-600'
                    : 'text-gray-600 border-transparent hover:text-blue-500 hover:border-blue-300'
                }`
              }
            >
              <Icon className="w-4 h-4 sm:w-5 sm:h-5" />
              <span className="hidden sm:inline">{label}</span>
            </NavLink>
          ))}
        </div>
      </div>
    </nav>
  );
}

export default Navigation;
