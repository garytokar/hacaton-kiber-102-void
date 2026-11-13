import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { BashkirOrnament } from '../components/ui/BashkirOrnament';
import { BookOpen, LogIn, Sparkles, AlertCircle } from 'lucide-react';

export const LoginPage: React.FC = () => {
  const { login, quickLogin } = useAuth();
  const navigate = useNavigate();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await login({ email, password });
      navigate('/dashboard');
    } catch (err: any) {
      const msg = typeof err === 'string' ? err : (err?.message && typeof err.message === 'string' ? err.message : (typeof err === 'object' ? JSON.stringify(err) : 'Ошибка входа'));
      setError(msg || 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  const handleQuickLogin = async (demoEmail: string) => {
    setLoading(true);
    setError(null);
    try {
      await quickLogin(demoEmail);
      navigate('/dashboard');
    } catch (err: any) {
      const msg = typeof err === 'string' ? err : (err?.message && typeof err.message === 'string' ? err.message : (typeof err === 'object' ? JSON.stringify(err) : 'Ошибка демо-входа'));
      setError(msg || 'Ошибка демо-входа');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-140px)] flex items-center justify-center p-4 select-none">
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-md max-w-md w-full p-8 space-y-6">
        <div className="text-center space-y-2">
          <div className="w-14 h-14 rounded-2xl bg-[#0C2340] flex items-center justify-center text-white mx-auto shadow-sm">
            <BashkirOrnament className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-black text-[#0C2340]">Вход в аккаунт</h2>
          <p className="text-xs text-slate-500">
            Войдите, чтобы продолжить путь Хранителя башкирского языка
          </p>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 text-xs p-3 rounded-xl flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="name@example.ru"
              className="w-full px-4 py-2.5 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1">Пароль</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-4 py-2.5 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-sm"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md transition-colors disabled:opacity-50"
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>
        </form>

        {/* Quick Demo Logins for Hackathon Judges */}
        <div className="pt-4 border-t border-slate-100">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider text-center mb-2.5 flex items-center justify-center gap-1">
            <Sparkles className="w-3.5 h-3.5 text-amber-500" />
            <span>Быстрый вход для жюри и тестов:</span>
          </div>

          <div className="grid grid-cols-3 gap-2 text-xs">
            <button
              type="button"
              onClick={() => handleQuickLogin('student_timur@yanlander.ru')}
              className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors"
            >
              Ученик
            </button>
            <button
              type="button"
              onClick={() => handleQuickLogin('teacher_aygul@yanlander.ru')}
              className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors"
            >
              Учитель
            </button>
            <button
              type="button"
              onClick={() => handleQuickLogin('expert@yanlander.ru')}
              className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors"
            >
              Эксперт
            </button>
          </div>
        </div>

        <div className="text-center text-xs text-slate-500">
          Ещё нет аккаунта?{' '}
          <Link to="/register" className="text-[#1E50A0] font-bold hover:underline">
            Зарегистрироваться
          </Link>
        </div>
      </div>
    </div>
  );
};
