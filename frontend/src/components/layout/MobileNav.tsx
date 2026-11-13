import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useLang } from '../../context/LanguageContext';
import { Home, BookOpen, Gem, User } from 'lucide-react';

export const MobileNav: React.FC = () => {
  const { user, openAuthModal } = useAuth();
  const { t } = useLang();

  const items = [
    { to: '/', label: t('nav_home'), icon: Home, protected: false },
    { to: '/task', label: t('nav_tasks'), icon: BookOpen, protected: false },
    { to: '/treasury', label: t('nav_treasury'), icon: Gem, protected: false },
    { to: '/profile', label: t('nav_profile'), icon: User, protected: true },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-white/95 backdrop-blur-md border-t border-slate-200 px-3 py-2 flex justify-around items-center md:hidden shadow-lg">
      {items.map((item) => {
        const Icon = item.icon;
        return (
          <NavLink
            key={item.to}
            to={item.to}
            onClick={(e) => {
              if (item.protected && !user) {
                e.preventDefault();
                openAuthModal(item.to);
              }
            }}
            className={({ isActive }) =>
              `flex flex-col items-center py-1 px-3 rounded-xl text-xs font-semibold transition-colors ${
                isActive ? 'text-[#1E50A0] font-bold' : 'text-slate-500 hover:text-slate-800'
              }`
            }
          >
            <Icon className="w-5 h-5 mb-0.5" />
            <span className="text-[11px] leading-tight">{item.label}</span>
          </NavLink>
        );
      })}
    </nav>
  );
};
