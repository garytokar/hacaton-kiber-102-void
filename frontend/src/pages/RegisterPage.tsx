import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { api } from '../api/client';
import { SchoolOption } from '../types';
import { BashkirOrnament } from '../components/ui/BashkirOrnament';
import { UserPlus, AlertCircle, ArrowLeft } from 'lucide-react';

export const RegisterPage: React.FC = () => {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [fullName, setFullName] = useState('');
  const [role, setRole] = useState('student');
  const [schoolId, setSchoolId] = useState<number | undefined>();
  const [classId, setClassId] = useState<number | undefined>();
  const [schools, setSchools] = useState<SchoolOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getSchools().then(setSchools).catch(() => {});
  }, []);

  const selectedSchool = schools.find((s) => s.id === schoolId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await register({
        email,
        username,
        password,
        full_name: fullName || username,
        role,
        school_id: schoolId,
        class_id: classId,
      });
      navigate('/dashboard');
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

  return (
    <div className="min-h-screen py-10 flex flex-col items-center justify-center p-4">
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-md max-w-lg w-full p-8 space-y-6">
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
          <h2 className="text-2xl font-black text-[#0C2340]">Регистрация Хранителя</h2>
          <p className="text-xs text-slate-500">
            Создайте аккаунт и присоединяйтесь к сообществу сохранения языка
          </p>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 text-xs p-3 rounded-xl flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">Никнейм *</label>
              <input
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="ural_batyr"
                className="w-full px-4 py-2 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">Имя и фамилия</label>
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                placeholder="Салават Юлаев"
                className="w-full px-4 py-2 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-sm"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1">Email *</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="name@example.ru"
              className="w-full px-4 py-2 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1">Пароль *</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-4 py-2 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-sm"
            />
          </div>

          {/* Role Choice */}
          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1.5">Роль участия:</label>
            <div className="grid grid-cols-3 gap-2">
              {[
                { id: 'student', label: 'Ученик' },
                { id: 'teacher', label: 'Учитель' },
                { id: 'expert', label: 'Эксперт' },
              ].map((r) => (
                <button
                  type="button"
                  key={r.id}
                  onClick={() => setRole(r.id)}
                  className={`py-2 px-3 rounded-full text-xs font-bold border transition-colors ${
                    role === r.id
                      ? 'bg-[#0C2340] text-white border-[#0C2340] shadow-sm'
                      : 'bg-slate-50 border-slate-200 text-slate-600 hover:bg-slate-100'
                  }`}
                >
                  {r.label}
                </button>
              ))}
            </div>
          </div>

          {/* School Selector */}
          <div>
            <label className="block text-xs font-bold text-slate-700 mb-1">
              Школа или учебное заведение:
            </label>
            <select
              value={schoolId || ''}
              onChange={(e) => {
                setSchoolId(e.target.value ? Number(e.target.value) : undefined);
                setClassId(undefined);
              }}
              className="w-full px-4 py-2 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-xs sm:text-sm"
            >
              <option value="">Индивидуальное участие (без школы)</option>
              {schools.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.city})
                </option>
              ))}
            </select>
          </div>

          {/* Class selector */}
          {selectedSchool && selectedSchool.classes.length > 0 && (
            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">Класс:</label>
              <select
                value={classId || ''}
                onChange={(e) => setClassId(e.target.value ? Number(e.target.value) : undefined)}
                className="w-full px-4 py-2 rounded-full bg-[#F1F5F9] border-none text-slate-900 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-xs sm:text-sm"
              >
                <option value="">Выберите класс</option>
                {selectedSchool.classes.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md transition-colors disabled:opacity-50"
          >
            {loading ? 'Создание...' : 'Зарегистрироваться'}
          </button>
        </form>

        <div className="text-center text-xs text-slate-500">
          Уже есть аккаунт?{' '}
          <Link to="/login" className="text-[#1E50A0] font-bold hover:underline">
            Войти
          </Link>
        </div>
      </div>
    </div>
  );
};
