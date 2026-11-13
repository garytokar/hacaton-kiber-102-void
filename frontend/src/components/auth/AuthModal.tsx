import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { BashkirOrnament } from '../ui/BashkirOrnament';
import { Sparkles, AlertCircle, X, LogIn, UserPlus } from 'lucide-react';

export const AuthModal: React.FC = () => {
  const {
    authModalOpen,
    authModalTab,
    authRedirectPath,
    closeAuthModal,
    setAuthModalTab,
    login,
    register,
    quickLogin
  } = useAuth();
  const navigate = useNavigate();

  // Login form state
  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');

  // Register form state
  const [regUsername, setRegUsername] = useState('');
  const [regFullName, setRegFullName] = useState('');
  const [regEmail, setRegEmail] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [regRole, setRegRole] = useState<'student' | 'teacher' | 'expert'>('student');

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Close on Escape key
  useEffect(() => {
    if (!authModalOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        closeAuthModal();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [authModalOpen]);

  if (!authModalOpen) return null;

  const handleSuccess = () => {
    closeAuthModal();
    if (authRedirectPath) {
      navigate(authRedirectPath);
    }
  };

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await login({ email: loginEmail, password: loginPassword });
      handleSuccess();
    } catch (err: any) {
      let msg = typeof err === 'string' ? err : (err?.message && typeof err.message === 'string' ? err.message : '');
      if (!msg || msg === '[object Object]') {
        msg = 'Неверный email или пароль. Попробуйте снова или воспользуйтесь быстрым входом.';
      }
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await register({
        username: regUsername.trim(),
        full_name: regFullName.trim() || undefined,
        email: regEmail.trim(),
        password: regPassword,
        role: regRole,
      });
      handleSuccess();
    } catch (err: any) {
      let msg = typeof err === 'string' ? err : (err?.message && typeof err.message === 'string' ? err.message : '');
      if (!msg || msg === '[object Object]') {
        msg = 'Ошибка регистрации. Проверьте правильность заполнения полей.';
      }
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  const handleQuick = async (demoEmail: string) => {
    setLoading(true);
    setError(null);
    try {
      await quickLogin(demoEmail);
      handleSuccess();
    } catch (err: any) {
      let msg = typeof err === 'string' ? err : (err?.message && typeof err.message === 'string' ? err.message : '');
      if (!msg || msg === '[object Object]') {
        msg = 'Ошибка демо-входа. Пожалуйста, попробуйте снова.';
      }
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4 animate-fade-in"
      onClick={(e) => {
        if (e.target === e.currentTarget) closeAuthModal();
      }}
    >
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-2xl max-w-md w-full p-6 sm:p-8 space-y-5 relative animate-scale-up">
        {/* Close button */}
        <button
          type="button"
          onClick={closeAuthModal}
          className="absolute top-5 right-5 w-8 h-8 rounded-full bg-slate-100 hover:bg-slate-200 text-slate-500 hover:text-slate-800 flex items-center justify-center transition-colors cursor-pointer"
          title="Закрыть"
        >
          <X className="w-4 h-4" />
        </button>

        {/* Brand Ornament Logo & Header */}
        <div className="text-center space-y-1.5 pt-1">
          <div className="w-12 h-12 rounded-2xl bg-[#0C2340] flex items-center justify-center text-white mx-auto shadow-sm">
            <BashkirOrnament className="w-7 h-7 text-white" />
          </div>

          <h3 className="text-xl sm:text-2xl font-black text-[#0C2340]">
            {authModalTab === 'login' ? 'Вход в аккаунт' : 'Регистрация Хранителя'}
          </h3>
          <p className="text-xs text-slate-500">
            {authModalTab === 'login'
              ? 'Войдите, чтобы продолжить путь Хранителя языка'
              : 'Создайте аккаунт и присоединяйтесь к сообществу'}
          </p>
        </div>

        {/* Tab Switcher Pills */}
        <div className="grid grid-cols-2 gap-1.5 p-1 bg-slate-100 rounded-2xl text-xs font-bold">
          <button
            type="button"
            onClick={() => {
              setAuthModalTab('login');
              setError(null);
            }}
            className={`py-2 rounded-xl flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
              authModalTab === 'login'
                ? 'bg-white text-[#0C2340] shadow-xs'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <LogIn className="w-3.5 h-3.5" />
            <span>Вход</span>
          </button>
          <button
            type="button"
            onClick={() => {
              setAuthModalTab('register');
              setError(null);
            }}
            className={`py-2 rounded-xl flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
              authModalTab === 'register'
                ? 'bg-white text-[#0C2340] shadow-xs'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <UserPlus className="w-3.5 h-3.5" />
            <span>Регистрация</span>
          </button>
        </div>

        {/* Error notification */}
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 text-xs p-3 rounded-xl flex items-center gap-2 animate-fade-in">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* LOGIN TAB */}
        {authModalTab === 'login' && (
          <div className="space-y-4">
            <form onSubmit={handleLoginSubmit} className="space-y-3.5">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Электронная почта
                </label>
                <input
                  type="email"
                  required
                  value={loginEmail}
                  onChange={(e) => setLoginEmail(e.target.value)}
                  placeholder="name@example.ru"
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Пароль
                </label>
                <input
                  type="password"
                  required
                  value={loginPassword}
                  onChange={(e) => setLoginPassword(e.target.value)}
                  placeholder="••••••••"
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold transition-all shadow-md cursor-pointer disabled:opacity-50"
              >
                {loading ? 'Вход...' : 'Войти'}
              </button>
            </form>

            {/* Quick Demo Logins */}
            <div className="pt-3 border-t border-slate-100">
              <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider text-center mb-2 flex items-center justify-center gap-1">
                <Sparkles className="w-3.5 h-3.5 text-amber-500" />
                <span>Быстрый вход для тестов:</span>
              </div>

              <div className="grid grid-cols-3 gap-2 text-xs">
                <button
                  type="button"
                  onClick={() => handleQuick('student_timur@yanlander.ru')}
                  className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors cursor-pointer"
                >
                  Ученик
                </button>
                <button
                  type="button"
                  onClick={() => handleQuick('teacher_aygul@yanlander.ru')}
                  className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors cursor-pointer"
                >
                  Учитель
                </button>
                <button
                  type="button"
                  onClick={() => handleQuick('expert@yanlander.ru')}
                  className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors cursor-pointer"
                >
                  Эксперт
                </button>
              </div>
            </div>

            <div className="text-center text-xs text-slate-500 pt-1">
              Ещё нет аккаунта?{' '}
              <button
                type="button"
                onClick={() => {
                  setAuthModalTab('register');
                  setError(null);
                }}
                className="text-[#1E50A0] font-bold hover:underline cursor-pointer"
              >
                Зарегистрироваться
              </button>
            </div>
          </div>
        )}

        {/* REGISTER TAB */}
        {authModalTab === 'register' && (
          <div className="space-y-4">
            <form onSubmit={handleRegisterSubmit} className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-bold text-slate-700 mb-1">Никнейм *</label>
                  <input
                    type="text"
                    required
                    value={regUsername}
                    onChange={(e) => setRegUsername(e.target.value)}
                    placeholder="ural_batyr"
                    className="w-full px-3.5 py-2 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                  />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-700 mb-1">Имя и фамилия</label>
                  <input
                    type="text"
                    value={regFullName}
                    onChange={(e) => setRegFullName(e.target.value)}
                    placeholder="Тимур Ишмухаметов"
                    className="w-full px-3.5 py-2 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Электронная почта *</label>
                <input
                  type="email"
                  required
                  value={regEmail}
                  onChange={(e) => setRegEmail(e.target.value)}
                  placeholder="name@example.ru"
                  className="w-full px-3.5 py-2 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Пароль *</label>
                <input
                  type="password"
                  required
                  value={regPassword}
                  onChange={(e) => setRegPassword(e.target.value)}
                  placeholder="Минимум 6 символов"
                  className="w-full px-3.5 py-2 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Роль на платформе</label>
                <select
                  value={regRole}
                  onChange={(e) => setRegRole(e.target.value as any)}
                  className="w-full px-3.5 py-2 rounded-xl border border-slate-300 text-xs text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] bg-white"
                >
                  <option value="student">Ученик / Хранитель</option>
                  <option value="teacher">Учитель</option>
                  <option value="expert">Эксперт-лингвист</option>
                </select>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold transition-all shadow-md cursor-pointer disabled:opacity-50 mt-1"
              >
                {loading ? 'Создание аккаунта...' : 'Зарегистрироваться'}
              </button>
            </form>

            <div className="text-center text-xs text-slate-500 pt-1">
              Уже зарегистрированы?{' '}
              <button
                type="button"
                onClick={() => {
                  setAuthModalTab('login');
                  setError(null);
                }}
                className="text-[#1E50A0] font-bold hover:underline cursor-pointer"
              >
                Войти
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
