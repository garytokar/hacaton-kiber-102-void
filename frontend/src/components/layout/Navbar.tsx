import React, { useState, useEffect, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useLang } from '../../context/LanguageContext';
import { UserAvatar } from '../ui/UserAvatar';
import {
  Search,
  Bell,
  ChevronDown,
  LogOut,
  Coins,
  Languages,
  User as UserIcon,
  LogIn
} from 'lucide-react';

export const Navbar: React.FC = () => {
  const { user, logout, openAuthModal } = useAuth();
  const { lang, toggleLang, t } = useLang();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [profileMenuOpen, setProfileMenuOpen] = useState(false);
  const profileMenuRef = useRef<HTMLDivElement>(null);

  // Close profile dropdown on click outside and on Escape
  useEffect(() => {
    if (!profileMenuOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(e.target as Node)) {
        setProfileMenuOpen(false);
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setProfileMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [profileMenuOpen]);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) {
      navigate(`/treasury?search=${encodeURIComponent(search.trim())}`);
    }
  };

  const displayName = user?.full_name || user?.username || '';
  const displayLevel = user?.stats?.level || 1;

  return (
    // COMPACT STATIC NAVBAR (Height ~48px, sleek and clean)
    <header className="relative z-20 bg-white border-b border-slate-200/80 px-4 sm:px-6 py-2 transition-all">
      <div className="flex items-center justify-between gap-4">
        {/* Left: Search input */}
        <form onSubmit={handleSearchSubmit} className="relative flex-1 max-w-md">
          <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-2.5" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('search_placeholder')}
            className="w-full pl-8 pr-4 py-1.5 rounded-full bg-[#F1F5F9] border-none text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-[#1E50A0] transition-all h-8"
          />
        </form>

        {/* Right Actions */}
        <div className="flex items-center gap-2 shrink-0">
          {/* Currency pill linking to shop (only when user logged in) */}
          {user && (
            <Link
              to="/shop"
              className="hidden lg:flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-50 hover:bg-amber-100 border border-amber-200/80 text-amber-800 text-xs font-semibold transition-all"
              title={t('nav_shop')}
            >
              <Coins className="w-3.5 h-3.5 text-amber-500" />
              <span>{user.stats?.currency ?? 0}</span>
              <span className="text-xs text-amber-600 font-semibold">{t('quray_currency')}</span>
            </Link>
          )}

          {/* Language Switcher */}
          <button
            onClick={toggleLang}
            className="flex items-center gap-1 px-2.5 py-1 rounded-full bg-[#EEF2FF] hover:bg-[#E0EAFF] text-[#1E50A0] text-xs font-bold transition-all cursor-pointer border border-blue-200/60"
            title="Переключить язык"
          >
            <Languages className="w-3.5 h-3.5" />
            <span>{lang === 'ru' ? 'РУС' : 'БАШ'}</span>
          </button>

          {/* Notifications bell circular button (clean, no fake blue dot, no alert) */}
          <button
            type="button"
            className="w-8 h-8 rounded-full bg-[#F1F5F9] hover:bg-slate-200 text-slate-600 flex items-center justify-center transition-colors cursor-pointer"
            title="Уведомления"
          >
            <Bell className="w-4 h-4 text-slate-500" />
          </button>

          {/* User profile pill or Guest Login button */}
          {user ? (
            <div className="relative" ref={profileMenuRef}>
              <button
                type="button"
                onClick={() => setProfileMenuOpen(!profileMenuOpen)}
                className="flex items-center gap-1.5 p-1 rounded-full hover:bg-slate-100 transition-colors cursor-pointer"
              >
                <UserAvatar user={user} size="sm" />
                <div className="text-left hidden sm:block">
                  <div className="text-xs font-semibold text-slate-900 leading-tight">
                    {displayName}
                  </div>
                  <div className="text-xs text-slate-400 font-normal">
                    {t('level')} {displayLevel}
                  </div>
                </div>
                <ChevronDown className="w-3.5 h-3.5 text-slate-400 hidden sm:block" />
              </button>

              {/* Profile Dropdown */}
              {profileMenuOpen && (
                <div className="absolute right-0 mt-2 w-52 bg-white rounded-2xl shadow-xl border border-slate-200 p-2 space-y-1 z-50 animate-fade-in text-xs">
                  <div className="px-3 py-2 border-b border-slate-100">
                    <div className="font-bold text-slate-900">{displayName}</div>
                    <div className="text-xs text-slate-400">{user.email}</div>
                  </div>

                  <Link
                    to="/profile"
                    onClick={() => setProfileMenuOpen(false)}
                    className="flex items-center gap-2 px-3 py-2 rounded-xl font-medium text-slate-700 hover:bg-slate-100"
                  >
                    <UserIcon className="w-4 h-4 text-slate-500" />
                    <span>{t('nav_profile')}</span>
                  </Link>

                  <Link
                    to="/shop"
                    onClick={() => setProfileMenuOpen(false)}
                    className="flex items-center gap-2 px-3 py-2 rounded-xl font-medium text-slate-700 hover:bg-slate-100"
                  >
                    <Coins className="w-4 h-4 text-amber-500" />
                    <span>{t('nav_shop')}</span>
                  </Link>

                  <div className="pt-1 border-t border-slate-100">
                    <button
                      type="button"
                      onClick={() => {
                        logout();
                        setProfileMenuOpen(false);
                      }}
                      className="w-full flex items-center gap-2 px-3 py-2 rounded-xl font-medium text-red-600 hover:bg-red-50 text-left"
                    >
                      <LogOut className="w-4 h-4" />
                      <span>{t('logout')}</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <button
              type="button"
              onClick={() => openAuthModal()}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-[#0C2340] hover:bg-[#1E50A0] text-white text-xs font-semibold transition-colors shadow-xs cursor-pointer"
            >
              <LogIn className="w-3.5 h-3.5" />
              <span>{t('login')}</span>
            </button>
          )}
        </div>
      </div>
    </header>
  );
};
