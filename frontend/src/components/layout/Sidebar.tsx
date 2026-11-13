import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { BashkirOrnament } from '../ui/BashkirOrnament';
import sidebarMountains from '../../assets/figma/sidebar_mountains.png';
import {
  Home,
  BookOpen,
  Gem,
  Trophy,
  Award,
  Users,
  BarChart2,
  BookA,
  User as UserIcon,
  Settings,
  ArrowRight
} from 'lucide-react';

export const Sidebar: React.FC = () => {
  const { user } = useAuth();

  const navItems = [
    { to: '/dashboard', label: 'Главная', icon: Home },
    { to: '/task', label: 'Обучение...', icon: BookOpen },
    { to: '/treasury', label: 'Сокровищни...', icon: Gem },
    { to: '/leaderboard', label: 'Рейтинг...', icon: Trophy },
    { to: '/achievements', label: 'Достижен...', icon: Award },
    { to: '/expeditions', label: 'Команда...', icon: Users },
    { to: '/profile#stats', label: 'Статистика', icon: BarChart2 },
    { to: '/dictionary', label: 'Словарь', icon: BookA },
    { to: '/profile', label: 'Профиль...', icon: UserIcon },
  ];

  if (user?.role === 'admin' || user?.role === 'expert' || user?.role === 'teacher') {
    navItems.push({ to: '/admin', label: 'Модерация...', icon: Settings });
  }

  return (
    <aside className="w-56 bg-[#0C2340] min-h-screen flex-col justify-between hidden md:flex shrink-0 select-none shadow-xl z-20">
      <div className="p-4 space-y-6">
        {/* Brand Logo */}
        <NavLink to="/" className="flex flex-col items-center gap-2 pt-2 group">
          <div className="text-white group-hover:scale-105 transition-transform">
            <BashkirOrnament className="w-10 h-10 text-white" />
          </div>
          <span className="font-extrabold text-sm tracking-widest text-white uppercase font-sans">
            ЙӘНЛӘНДЕР
          </span>
        </NavLink>

        {/* Navigation list */}
        <nav className="space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-xs font-semibold transition-all ${
                    isActive
                      ? 'bg-[#1E50A0] text-white shadow-md'
                      : 'text-slate-300 hover:text-white hover:bg-white/10'
                  }`
                }
              >
                <Icon className="w-4 h-4 shrink-0" />
                <span className="truncate">{item.label}</span>
              </NavLink>
            );
          })}
        </nav>
      </div>

      {/* Bottom Mountain Graphic Callout */}
      <div className="relative pt-4 overflow-hidden">
        <img
          src={sidebarMountains}
          alt="Уральские горы"
          className="w-full h-auto object-cover opacity-90"
        />
        <div className="p-4 bg-gradient-to-t from-[#0C2340] via-[#0C2340]/90 to-transparent -mt-10 relative z-10 flex items-center justify-between">
          <div className="text-[11px] font-bold text-white leading-tight">
            Сохраняй <br />
            башкирский язык
          </div>
          <NavLink
            to="/task"
            className="w-7 h-7 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white flex items-center justify-center transition-transform hover:scale-110 shadow"
            title="Перейти к заданию"
          >
            <ArrowRight className="w-3.5 h-3.5" />
          </NavLink>
        </div>
      </div>
    </aside>
  );
};
