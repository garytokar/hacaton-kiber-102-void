import React from 'react';
import { NavLink } from 'react-router-dom';
import { LayoutDashboard, CheckCircle2, Gem, BookA, Trophy, User } from 'lucide-react';

export const MobileNav: React.FC = () => {
  const items = [
    { to: '/dashboard', label: 'Главная', icon: LayoutDashboard },
    { to: '/task', label: 'Задание', icon: CheckCircle2 },
    { to: '/treasury', label: 'Сокровища', icon: Gem },
    { to: '/dictionary', label: 'Словарь', icon: BookA },
    { to: '/leaderboard', label: 'Рейтинг', icon: Trophy },
    { to: '/profile', label: 'Профиль', icon: User },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-white/95 backdrop-blur border-t border-[#E2E8F0] px-2 py-1.5 flex justify-around items-center lg:hidden">
      {items.map((item) => {
        const Icon = item.icon;
        return (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `flex flex-col items-center py-1 px-2 rounded-lg text-[10px] font-semibold transition-colors ${
                isActive ? 'text-[#1E50A0] font-bold' : 'text-slate-400 hover:text-slate-600'
              }`
            }
          >
            <Icon className="w-5 h-5 mb-0.5" />
            <span>{item.label}</span>
          </NavLink>
        );
      })}
    </nav>
  );
};
