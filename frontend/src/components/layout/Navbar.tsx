import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import defaultAvatar from '../../assets/figma/avatar_user.png';
import {
  Search,
  Palette,
  Bell,
  ChevronDown,
  LogOut,
  Sparkles,
  Flame,
  Coins,
  Shield,
  User as UserIcon
} from 'lucide-react';

export const Navbar: React.FC = () => {
  const { user, logout, quickLogin } = useAuth();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [profileMenuOpen, setProfileMenuOpen] = useState(false);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) {
      navigate(`/treasury?search=${encodeURIComponent(search.trim())}`);
    }
  };

  return (
    <header className="sticky top-0 z-30 bg-white border-b border-slate-200/80 px-4 lg:px-8 py-3">
      <div className="flex items-center justify-between gap-4">
        {/* Left: Search pill */}
        <form onSubmit={handleSearchSubmit} className="relative flex-1 max-w-md">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-2.5" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск слов и переводов..."
            className="w-full pl-10 pr-4 py-2 rounded-full bg-[#F1F5F9] border-none text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] transition-all"
          />
        </form>

        {/* Right Actions */}
        <div className="flex items-center gap-3 sm:gap-4">
          {/* Color palette pill */}
          <button
            onClick={() => alert('Цветовая схема: Синяя палитра (Дизайн хакатона)')}
            className="hidden sm:flex items-center gap-1.5 px-3.5 py-1.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] hover:bg-[#E0EAFF] text-xs font-bold transition-colors"
          >
            <Palette className="w-3.5 h-3.5" />
            <span>Цветовая модель</span>
          </button>

          {/* Notifications bell with badge */}
          <button
            onClick={() => alert('Уведомлений нет')}
            className="relative p-2 text-slate-500 hover:text-slate-800 hover:bg-slate-100 rounded-full transition-colors"
            title="Уведомления"
          >
            <Bell className="w-4 h-4" />
            <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-amber-500 ring-2 ring-white" />
          </button>

          {/* User profile / Auth */}
          {user ? (
            <div className="relative">
              <button
                onClick={() => setProfileMenuOpen(!profileMenuOpen)}
                className="flex items-center gap-2.5 p-1 rounded-full hover:bg-slate-100 transition-colors"
              >
                <img
                  src={defaultAvatar}
                  alt={user.full_name || user.username}
                  className="w-8 h-8 rounded-full object-cover border border-slate-200 shadow-sm"
                  onError={(e) => {
                    // Fallback to stylized initials if image fails
                    (e.target as HTMLElement).style.display = 'none';
                  }}
                />
                <div className="text-left hidden sm:block">
                  <div className="text-xs font-bold text-slate-900 leading-tight">
                    {user.full_name || user.username}
                  </div>
                  <div className="text-[10px] text-slate-400 font-medium">
                    Уровень {user.stats?.level || 1}
                  </div>
                </div>
                <ChevronDown className="w-3.5 h-3.5 text-slate-400 hidden sm:block" />
              </button>

              {/* Profile Dropdown */}
              {profileMenuOpen && (
                <div className="absolute right-0 mt-2 w-56 bg-white rounded-2xl shadow-xl border border-slate-200 p-2 space-y-1 z-50 animate-fade-in">
                  <div className="px-3 py-2 border-b border-slate-100">
                    <div className="text-xs font-bold text-slate-900">{user.full_name || user.username}</div>
                    <div className="text-[11px] text-slate-400">{user.email}</div>
                    <div className="text-[10px] font-bold text-[#1E50A0] mt-1 capitalize">Роль: {user.role}</div>
                  </div>

                  <Link
                    to="/profile"
                    onClick={() => setProfileMenuOpen(false)}
                    className="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-semibold text-slate-700 hover:bg-slate-100"
                  >
                    <UserIcon className="w-4 h-4" />
                    <span>Мой профиль</span>
                  </Link>

                  <div className="pt-1 border-t border-slate-100">
                    <button
                      onClick={() => {
                        setProfileMenuOpen(false);
                        logout();
                        navigate('/');
                      }}
                      className="w-full flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-semibold text-red-600 hover:bg-red-50"
                    >
                      <LogOut className="w-4 h-4" />
                      <span>Выйти</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <Link
                to="/login"
                className="px-4 py-1.5 rounded-full text-xs font-bold text-[#0C2340] hover:bg-slate-100 transition-colors"
              >
                Войти
              </Link>
              <Link
                to="/register"
                className="px-4 py-1.5 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold transition-all shadow-sm"
              >
                Регистрация
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};
