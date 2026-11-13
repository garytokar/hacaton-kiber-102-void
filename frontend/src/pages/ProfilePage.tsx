import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { ActivityPoint } from '../types';
import defaultAvatar from '../assets/figma/avatar_user.png';
import {
  User, Shield, Flame, Coins, Sparkles, School, Edit2, Check,
  Calendar, CheckCircle2, Trophy, Clock
} from 'lucide-react';

export const ProfilePage: React.FC = () => {
  const { user, refreshUser } = useAuth();
  const [activity, setActivity] = useState<ActivityPoint[]>([]);
  const [isEditing, setIsEditing] = useState(false);
  const [fullName, setFullName] = useState(user?.full_name || '');
  const [protecting, setProtecting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    api.getUserActivity().then(setActivity).catch(console.error);
  }, []);

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.updateProfile({ full_name: fullName });
      setIsEditing(false);
      await refreshUser();
    } catch (err: any) {
      alert(err.message || 'Ошибка обновления');
    }
  };

  const handleBuyProtection = async () => {
    setProtecting(true);
    setMessage(null);
    try {
      const res = await api.protectStreak();
      setMessage(res.message);
      await refreshUser();
    } catch (err: any) {
      alert(err.message || 'Ошибка покупки');
    } finally {
      setProtecting(false);
    }
  };

  const stats = user?.stats;

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16 select-none">
      {/* Profile Header Card */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <img
              src={defaultAvatar}
              alt={user?.full_name || user?.username}
              className="w-16 h-16 rounded-full object-cover border-2 border-slate-200 shadow-sm"
              onError={(e) => {
                (e.target as HTMLElement).style.display = 'none';
              }}
            />

            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl sm:text-2xl font-black text-[#0C2340]">
                  {user?.full_name || user?.username}
                </h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] uppercase">
                  {user?.role}
                </span>
              </div>
              <div className="text-xs text-slate-400 mt-0.5">
                @{user?.username} • {user?.email}
              </div>
              <div className="text-xs text-[#1E50A0] font-semibold mt-1">
                {user?.school_name || 'Индивидуальный участник'} {user?.class_name ? `(${user.class_name})` : ''}
              </div>
            </div>
          </div>

          <button
            onClick={() => setIsEditing(!isEditing)}
            className="px-4 py-2 rounded-full border border-slate-200 text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-1.5 self-start sm:self-auto"
          >
            <Edit2 className="w-3.5 h-3.5" />
            <span>{isEditing ? 'Отмена' : 'Редактировать'}</span>
          </button>
        </div>

        {/* Edit Form */}
        {isEditing && (
          <form onSubmit={handleSaveProfile} className="bg-slate-50 p-4 rounded-2xl border border-slate-200 flex flex-col sm:flex-row gap-3">
            <input
              type="text"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              placeholder="Полное имя"
              className="flex-1 px-4 py-2 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-[#1E50A0] bg-white"
            />
            <button
              type="submit"
              className="px-5 py-2 rounded-full bg-[#0C2340] text-white text-xs font-bold hover:bg-[#1A4480]"
            >
              Сохранить
            </button>
          </form>
        )}

        {/* Key Stats Strip */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-4 border-t border-slate-100">
          <div className="p-3.5 rounded-2xl bg-[#F8FAFC]">
            <span className="text-[11px] font-bold text-slate-400 uppercase">Уровень</span>
            <div className="text-base sm:text-lg font-black text-[#0C2340] mt-0.5">
              {stats?.level || 1} ({stats?.level_title || 'Хранитель'})
            </div>
          </div>

          <div className="p-3.5 rounded-2xl bg-[#F8FAFC]">
            <span className="text-[11px] font-bold text-slate-400 uppercase">Опыт Хранителя</span>
            <div className="text-base sm:text-lg font-black text-[#1E50A0] mt-0.5">
              {stats?.xp || 0} XP
            </div>
          </div>

          <div className="p-3.5 rounded-2xl bg-[#F8FAFC]">
            <span className="text-[11px] font-bold text-slate-400 uppercase">Серия активности</span>
            <div className="text-base sm:text-lg font-black text-amber-600 mt-0.5 flex items-center gap-1">
              <Flame className="w-4 h-4 fill-amber-500" />
              <span>{stats?.current_streak || 0} дн.</span>
            </div>
          </div>

          <div className="p-3.5 rounded-2xl bg-[#F8FAFC]">
            <span className="text-[11px] font-bold text-slate-400 uppercase">Баланс Ҡурай</span>
            <div className="text-base sm:text-lg font-black text-amber-600 mt-0.5 flex items-center gap-1">
              <Coins className="w-4 h-4 fill-amber-500" />
              <span>{stats?.currency || 0}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Streak Protection Store Item */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center shrink-0">
            <Shield className="w-6 h-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-extrabold text-sm text-[#0C2340]">
                Оберег непрерывной серии (Streak Freeze)
              </h3>
              {stats?.streak_protected && (
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-blue-100 text-[#1E50A0]">
                  Активен
                </span>
              )}
            </div>
            <p className="text-xs text-slate-500 mt-0.5">
              Защищает вашу серию от сгорания, если вы пропустите один день занятий.
            </p>
          </div>
        </div>

        <button
          onClick={handleBuyProtection}
          disabled={protecting || stats?.streak_protected || (stats?.currency || 0) < 50}
          className="px-5 py-2.5 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-extrabold text-xs shadow-md transition-all disabled:opacity-50 shrink-0"
        >
          {stats?.streak_protected ? 'Защита уже активна' : 'Активировать за 50 Ҡурай'}
        </button>
      </div>

      {message && (
        <div className="bg-blue-50 border border-blue-200 text-[#1E50A0] text-xs p-3 rounded-2xl flex items-center gap-2 font-semibold">
          <CheckCircle2 className="w-4 h-4" />
          <span>{message}</span>
        </div>
      )}

      {/* GitHub Contributions Style Activity Calendar */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Calendar className="w-5 h-5 text-[#1E50A0]" />
            <h3 className="font-extrabold text-base text-[#0C2340]">
              Календарь активности (последние 60 дней)
            </h3>
          </div>
          <span className="text-xs text-slate-400">Каждый день имеет значение</span>
        </div>

        {/* Heatmap Grid */}
        <div className="overflow-x-auto pb-2">
          <div className="flex gap-1.5 min-w-[500px]">
            {activity.map((pt) => {
              const count = pt.count;
              const bg =
                count === 0
                  ? 'bg-slate-100'
                  : count <= 2
                  ? 'bg-[#EEF2FF]'
                  : count <= 5
                  ? 'bg-[#1E50A0]'
                  : 'bg-[#0C2340]';

              return (
                <div
                  key={pt.date}
                  title={`${pt.date}: ${count} действий`}
                  className={`w-3.5 h-9 rounded-sm ${bg} hover:ring-2 hover:ring-blue-400 transition-all cursor-pointer`}
                />
              );
            })}
          </div>
        </div>

        <div className="flex items-center justify-between text-[11px] text-slate-400 pt-2 border-t border-slate-100">
          <span>Всего ответов и правок за период: {activity.reduce((acc, c) => acc + c.count, 0)}</span>
          <div className="flex items-center gap-1.5">
            <span>Меньше</span>
            <div className="w-3 h-3 rounded-xs bg-slate-100" />
            <div className="w-3 h-3 rounded-xs bg-[#EEF2FF]" />
            <div className="w-3 h-3 rounded-xs bg-[#1E50A0]" />
            <div className="w-3 h-3 rounded-xs bg-[#0C2340]" />
            <span>Больше</span>
          </div>
        </div>
      </div>
    </div>
  );
};
