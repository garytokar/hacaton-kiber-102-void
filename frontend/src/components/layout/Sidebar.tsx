import React, { useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useLang } from '../../context/LanguageContext';
import { BashkirOrnament } from '../ui/BashkirOrnament';
import {
  PanelsTopLeft,
  Languages,
  Archive,
  LibraryBig,
  CircleUserRound,
  Medal,
  Sparkles,
  Compass,
  ChartNoAxesCombined,
  Store,
  ShieldCheck,
  ChevronRight,
  ChevronLeft,
  ChevronDown,
  MoreHorizontal
} from 'lucide-react';

export const Sidebar: React.FC = () => {
  const { user, openAuthModal } = useAuth();
  const { t, lang } = useLang();
  const location = useLocation();

  // Controlled open/collapsed state
  const [isOpen, setIsOpen] = useState<boolean>(() => {
    try {
      const saved = localStorage.getItem('yanlander_sidebar_open');
      if (saved !== null) {
        return saved === 'true';
      }
      return true;
    } catch {
      return true;
    }
  });

  const toggleSidebar = () => {
    setIsOpen((prev) => {
      const next = !prev;
      try {
        localStorage.setItem('yanlander_sidebar_open', String(next));
      } catch {}
      return next;
    });
  };

  // 1. Primary items in first row: Home, Practice, Treasury, Dictionary, Profile
  const primaryNavItems = [
    { to: '/', labelKey: 'nav_home', icon: PanelsTopLeft },
    { to: '/task', labelKey: 'nav_tasks', icon: Languages },
    { to: '/treasury', labelKey: 'nav_treasury', icon: Archive },
    { to: '/dictionary', labelKey: 'nav_dictionary', icon: LibraryBig },
    { to: '/profile', labelKey: 'nav_profile', icon: CircleUserRound },
  ];

  // 2. Secondary items in "Ещё" (collapsed by default): Leaderboard, Achievements, Expeditions, Analytics, Shop
  const secondaryNavItems = [
    { to: '/leaderboard', labelKey: 'nav_leaderboard', icon: Medal },
    { to: '/achievements', labelKey: 'nav_achievements', icon: Sparkles },
    { to: '/expeditions', labelKey: 'nav_team', icon: Compass },
    { to: '/analytics', labelKey: 'nav_analytics', icon: ChartNoAxesCombined },
    { to: '/shop', labelKey: 'nav_shop', icon: Store },
  ];

  const isAnySecondaryActive = secondaryNavItems.some((item) => location.pathname === item.to);
  const [moreOpen, setMoreOpen] = useState<boolean>(isAnySecondaryActive);

  // Admin / moderation item
  const showAdmin = user?.role === 'admin' || user?.role === 'expert' || user?.role === 'teacher';

  const protectedPaths = ['/dictionary', '/profile', '/achievements', '/expeditions', '/shop', '/admin'];

  const renderLink = (item: { to: string; labelKey: string; icon: any }) => {
    const Icon = item.icon;
    const label = t(item.labelKey);
    const isProtected = protectedPaths.includes(item.to);

    return (
      <NavLink
        key={item.to}
        to={item.to}
        onClick={(e) => {
          if (isProtected && !user) {
            e.preventDefault();
            openAuthModal(item.to);
          }
        }}
        className={({ isActive }) =>
          isOpen
            ? `flex items-center gap-3 px-3 py-2.5 rounded-xl text-xs font-semibold transition-colors ${
                isActive
                  ? 'bg-[#1E50A0] text-white shadow-xs'
                  : 'text-slate-300 hover:text-white hover:bg-white/10'
              }`
            : `w-11 h-11 rounded-2xl mx-auto flex items-center justify-center transition-colors group relative ${
                isActive
                  ? 'bg-[#1E50A0] text-white shadow-xs ring-1 ring-blue-400/30'
                  : 'text-slate-400 hover:text-white hover:bg-white/10'
              }`
        }
        title={!isOpen ? label : undefined}
      >
        <Icon className={`${isOpen ? 'w-4 h-4' : 'w-5 h-5'} shrink-0`} />
        {isOpen ? (
          <span className="truncate">{label}</span>
        ) : (
          <span className="absolute left-14 bg-[#0C2340] text-white text-xs font-semibold px-2.5 py-1 rounded-md opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity whitespace-nowrap shadow-lg border border-slate-700/80 z-50">
            {label}
          </span>
        )}
      </NavLink>
    );
  };

  return (
    <aside
      className={`relative bg-[#0C2340] min-h-screen hidden md:flex flex-col justify-between py-4 select-none shadow-2xl z-30 transition-all duration-200 ease-in-out border-r border-[#15345b]/60 shrink-0 ${
        isOpen ? 'w-64 px-3' : 'w-20 items-center px-2'
      }`}
    >
      {/* Top Header & Navigation */}
      <div className="flex flex-col gap-4 w-full">
        {isOpen ? (
          /* EXPANDED HEADER */
          <div className="flex items-center justify-between px-2 pt-1 pb-2 border-b border-white/10">
            <NavLink
              to="/"
              className="flex items-center gap-3 group"
              title="ЙӘНЛӘНДЕР • На главную"
            >
              <div className="w-10 h-10 rounded-2xl bg-white/10 group-hover:bg-white/20 flex items-center justify-center text-white transition-all shadow-xs shrink-0">
                <BashkirOrnament className="w-6 h-6 text-white group-hover:scale-105 transition-transform" />
              </div>
              <div className="flex flex-col overflow-hidden">
                <span className="font-black text-sm tracking-wider text-white uppercase font-sans whitespace-nowrap leading-tight">
                  ЙӘНЛӘНДЕР
                </span>
                <span className="text-xs text-slate-400 font-medium whitespace-nowrap leading-tight">
                  Хранители языка
                </span>
              </div>
            </NavLink>

            {/* СТРЕЛКА ЗАКРЫТИЯ МЕНЮ */}
            <button
              type="button"
              onClick={toggleSidebar}
              className="w-8 h-8 rounded-xl bg-white/10 hover:bg-white/20 text-slate-300 hover:text-white flex items-center justify-center transition-colors cursor-pointer shrink-0"
              title="Свернуть меню"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
          </div>
        ) : (
          /* COLLAPSED HEADER */
          <div className="flex flex-col items-center gap-2 w-full pt-1">
            <NavLink
              to="/"
              className="w-11 h-11 rounded-2xl bg-white/10 hover:bg-white/20 flex items-center justify-center text-white transition-all shadow-xs group"
              title="ЙӘНЛӘНДЕР • На главную"
            >
              <BashkirOrnament className="w-6 h-6 text-white group-hover:scale-105 transition-transform" />
            </NavLink>

            {/* СТРЕЛКА ОТКРЫТИЯ МЕНЮ */}
            <button
              type="button"
              onClick={toggleSidebar}
              className="w-11 h-7 rounded-xl bg-white/10 hover:bg-[#1E50A0] text-slate-300 hover:text-white flex items-center justify-center transition-all cursor-pointer group shadow-xs"
              title="Развернуть меню"
            >
              <ChevronRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
            </button>
          </div>
        )}

        {/* Primary Navigation items */}
        <nav className="flex flex-col gap-1.5 w-full">
          {primaryNavItems.map(renderLink)}

          {/* "Ещё" section */}
          {isOpen ? (
            <div className="pt-1">
              <button
                type="button"
                onClick={() => setMoreOpen(!moreOpen)}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-semibold transition-colors cursor-pointer ${
                  isAnySecondaryActive ? 'text-blue-300 bg-white/5' : 'text-slate-400 hover:text-white hover:bg-white/10'
                }`}
              >
                <div className="flex items-center gap-3">
                  <MoreHorizontal className="w-4 h-4 shrink-0" />
                  <span>{lang === 'ba' ? 'Тағы' : 'Ещё'}</span>
                </div>
                <ChevronDown className={`w-3.5 h-3.5 transition-transform duration-200 ${moreOpen ? 'rotate-180' : ''}`} />
              </button>

              {moreOpen && (
                <div className="flex flex-col gap-1 pt-1 pl-1 space-y-1 animate-fade-in">
                  {secondaryNavItems.map(renderLink)}
                </div>
              )}
            </div>
          ) : (
            <div>
              <button
                type="button"
                onClick={() => setMoreOpen(!moreOpen)}
                className={`w-11 h-11 rounded-2xl mx-auto flex items-center justify-center transition-colors group relative cursor-pointer ${
                  moreOpen || isAnySecondaryActive
                    ? 'bg-white/15 text-white shadow-xs'
                    : 'text-slate-400 hover:text-white hover:bg-white/10'
                }`}
                title={lang === 'ba' ? 'Тағы' : 'Ещё'}
              >
                <MoreHorizontal className="w-5 h-5 shrink-0" />
                <span className="absolute left-14 bg-[#0C2340] text-white text-xs font-semibold px-2.5 py-1 rounded-md opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity whitespace-nowrap shadow-lg border border-slate-700/80 z-50">
                  {lang === 'ba' ? 'Тағы' : 'Ещё'}
                </span>
              </button>

              {moreOpen && (
                <div className="flex flex-col gap-1.5 pt-1.5 animate-fade-in">
                  {secondaryNavItems.map(renderLink)}
                </div>
              )}
            </div>
          )}

          {/* Moderation / Admin item */}
          {showAdmin && renderLink({ to: '/admin', labelKey: 'nav_admin', icon: ShieldCheck })}
        </nav>
      </div>

      {/* Bottom spacer (No duplicate Treasury button) */}
      <div className="w-full" />
    </aside>
  );
};
