import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { BashkirOrnament } from '../components/ui/BashkirOrnament';
import { BookOpen, LogIn, Sparkles, AlertCircle, ArrowLeft } from 'lucide-react';

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
      let msg = typeof err === 'string' ? err : (err?.message && typeof err.message === 'string' ? err.message : '');
      if (!msg || msg === '[object Object]') {
        msg = 'Неверный email или пароль. Проверьте введённые данные или используйте кнопки быстрого входа ниже.';
      }
      setError(msg);
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
    <div className="min-h-screen py-10 flex flex-col items-center justify-center p-4">
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-md max-w-md w-full p-8 space-y-6">
        {/* Link back to Home */}
        <div className="flex items-center justify-between pb-1">
          <Link
            to="/"
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-[#0C2340] transition-colors"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>На главную</span>
          </Link>
        </div>

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
            <label className="block text-xs font-bold text-slate-700 mb-1">
              Электронная почта
            </label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="name@example.ru"
              className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-xs focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1">
              Пароль
            </label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-xs focus:outline-none focus:ring-2 focus:ring-[#1E50A0]"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold transition-all shadow-md disabled:opacity-50"
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>
        </form>

        {/* Quick Demo Logins for Hackathon Judges */}
        <div className="pt-4 border-t border-slate-100">
          <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider text-center mb-2.5 flex items-center justify-center gap-1">
            <Sparkles className="w-3.5 h-3.5 text-amber-500" />
            <span>Быстрый вход для тестов:</span>
          </div>

          <div className="grid grid-cols-3 gap-2 text-xs">
            <button
              type="button"
              onClick={() => handleQuickLogin('student_timur@yanlander.ru')}
              className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors cursor-pointer"
            >
              Ученик
            </button>
            <button
              type="button"
              onClick={() => handleQuickLogin('teacher_aygul@yanlander.ru')}
              className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors cursor-pointer"
            >
              Учитель
            </button>
            <button
              type="button"
              onClick={() => handleQuickLogin('expert@yanlander.ru')}
              className="p-2 rounded-xl bg-slate-50 hover:bg-blue-50 border border-slate-200 text-slate-700 hover:text-[#1E50A0] text-center font-bold transition-colors cursor-pointer"
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
