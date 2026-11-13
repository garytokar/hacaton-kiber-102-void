import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { LeaderboardUser, TeamLeaderboardEntry, SchoolLeaderboardEntry } from '../types';
import {
  Trophy, Medal, Users, School, Flame, Sparkles, CheckCircle2, Shield, ArrowRight
} from 'lucide-react';

export const LeaderboardPage: React.FC = () => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'individual' | 'teams' | 'schools'>('individual');
  const [metric, setMetric] = useState<'xp' | 'quality'>('xp');
  const [period, setPeriod] = useState<'all_time' | 'monthly' | 'weekly'>('all_time');

  const [users, setUsers] = useState<LeaderboardUser[]>([]);
  const [teams, setTeams] = useState<TeamLeaderboardEntry[]>([]);
  const [schools, setSchools] = useState<SchoolLeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    if (activeTab === 'individual') {
      api.getLeaderboard(period, metric)
        .then(setUsers)
        .catch(console.error)
        .finally(() => setLoading(false));
    } else if (activeTab === 'teams') {
      api.getTeamLeaderboard()
        .then(setTeams)
        .catch(console.error)
        .finally(() => setLoading(false));
    } else {
      api.getSchoolLeaderboard()
        .then(setSchools)
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [activeTab, metric, period]);

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md border border-slate-700/50">
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
          <Trophy className="w-3.5 h-3.5 text-amber-400" />
          <span>Зал славы Хранителей</span>
        </div>
        <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
          Рейтинг Хранителей башкирского языка
        </h1>
        <p className="text-slate-300 text-xs sm:text-sm mt-1">
          Соревнуйтесь индивидуально или прославляйте свою школу и команду!
        </p>
      </div>

      {/* Tabs & Filters */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-white p-2.5 rounded-2xl border border-slate-200/90 shadow-sm">
        <div className="flex items-center gap-1.5 w-full sm:w-auto">
          <button
            onClick={() => setActiveTab('individual')}
            className={`flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-bold transition-all ${
              activeTab === 'individual'
                ? 'bg-[#1E50A0] text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-100'
            }`}
          >
            Пользователи
          </button>
          <button
            onClick={() => setActiveTab('schools')}
            className={`flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-bold transition-all ${
              activeTab === 'schools'
                ? 'bg-[#1E50A0] text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-100'
            }`}
          >
            Школы
          </button>
          <button
            onClick={() => setActiveTab('teams')}
            className={`flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-bold transition-all ${
              activeTab === 'teams'
                ? 'bg-[#1E50A0] text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-100'
            }`}
          >
            Команды
          </button>
        </div>

        {activeTab === 'individual' && (
          <div className="flex items-center gap-2 self-end sm:self-auto text-xs">
            <span className="text-slate-400 font-medium">Метрика:</span>
            <button
              onClick={() => setMetric('xp')}
              className={`px-3 py-1.5 rounded-lg font-bold transition-colors ${
                metric === 'xp' ? 'bg-[#EEF2FF] text-[#1E50A0]' : 'text-slate-500 hover:bg-slate-100'
              }`}
            >
              Опыт (XP)
            </button>
            <button
              onClick={() => setMetric('quality')}
              className={`px-3 py-1.5 rounded-lg font-bold transition-colors ${
                metric === 'quality' ? 'bg-[#EEF2FF] text-[#1E50A0]' : 'text-slate-500 hover:bg-slate-100'
              }`}
            >
              Точность
            </button>
          </div>
        )}
      </div>

      {/* Leaderboard Tables */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm overflow-hidden">
        {loading ? (
          <div className="text-center py-12 text-slate-400 text-sm">
            Загрузка таблицы рейтинга...
          </div>
        ) : activeTab === 'individual' ? (
          <div className="divide-y divide-slate-100">
            {users.map((u) => {
              const isCurrentUser = user?.id === u.user_id;
              return (
                <div
                  key={u.user_id}
                  className={`p-4 sm:px-6 flex items-center justify-between gap-4 transition-colors ${
                    isCurrentUser ? 'bg-blue-50/70 font-semibold' : 'hover:bg-slate-50'
                  }`}
                >
                  <div className="flex items-center gap-3 sm:gap-4 min-w-0">
                    {/* Rank Medal */}
                    <div className="w-8 text-center shrink-0">
                      {u.rank === 1 ? (
                        <span className="text-xl">🥇</span>
                      ) : u.rank === 2 ? (
                        <span className="text-xl">🥈</span>
                      ) : u.rank === 3 ? (
                        <span className="text-xl">🥉</span>
                      ) : (
                        <span className="text-xs font-bold text-slate-400">#{u.rank}</span>
                      )}
                    </div>

                    {/* Avatar / Initials */}
                    <div className="w-10 h-10 rounded-full bg-[#0C2340] text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-sm">
                      {u.full_name?.slice(0, 2) || u.username.slice(0, 2).toUpperCase()}
                    </div>

                    {/* Name & School */}
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-extrabold text-sm text-[#0C2340] truncate">
                          {u.full_name || u.username}
                        </span>
                        {isCurrentUser && (
                          <span className="text-[10px] bg-[#1E50A0] text-white px-2 py-0.5 rounded-full font-bold">
                            Вы
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-slate-400 truncate">
                        Уровень {u.level} • {u.school_name || 'Индивидуальный Хранитель'}
                      </div>
                    </div>
                  </div>

                  {/* Right: XP / Score */}
                  <div className="flex items-center gap-4 shrink-0">
                    {u.streak > 0 && (
                      <div className="hidden sm:flex items-center gap-1 text-xs font-bold text-amber-600 bg-amber-50 px-2.5 py-1 rounded-full">
                        <Flame className="w-3.5 h-3.5 fill-amber-500 text-amber-500" />
                        <span>{u.streak} дн.</span>
                      </div>
                    )}

                    <div className="text-right">
                      <div className="text-sm font-extrabold text-[#1E50A0]">
                        {`${u.xp.toLocaleString()} XP`}
                      </div>
                      <div className="text-[10px] text-slate-400 font-medium">
                        {u.confirmed_count} проверено
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        ) : activeTab === 'schools' ? (
          <div className="divide-y divide-slate-100">
            {schools.map((s, idx) => (
              <div
                key={s.school_id}
                className="p-4 sm:px-6 flex items-center justify-between gap-4 hover:bg-slate-50 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 text-center shrink-0">
                    {idx === 0 ? '🥇' : idx === 1 ? '🥈' : idx === 2 ? '🥉' : `#${idx + 1}`}
                  </div>
                  <div className="w-10 h-10 rounded-2xl bg-blue-100 text-[#1E50A0] flex items-center justify-center font-bold text-sm shrink-0">
                    <School className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="font-extrabold text-sm text-[#0C2340]">{s.name}</div>
                    <div className="text-xs text-slate-400">{s.students_count} активных учеников</div>
                  </div>
                </div>

                <div className="text-right">
                  <div className="text-sm font-extrabold text-[#1E50A0]">{s.total_xp.toLocaleString()} XP</div>
                  <div className="text-[10px] text-slate-400">{s.total_confirmed} фраз проверено</div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="divide-y divide-slate-100">
            {teams.map((t, idx) => (
              <div
                key={t.team_id}
                className="p-4 sm:px-6 flex items-center justify-between gap-4 hover:bg-slate-50 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 text-center shrink-0">
                    {idx === 0 ? '🥇' : idx === 1 ? '🥈' : idx === 2 ? '🥉' : `#${idx + 1}`}
                  </div>
                  <div className="w-10 h-10 rounded-2xl bg-indigo-100 text-[#0C2340] flex items-center justify-center font-bold text-sm shrink-0">
                    <Users className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="font-extrabold text-sm text-[#0C2340]">{t.name}</div>
                    <div className="text-xs text-slate-400">{t.members_count} участников</div>
                  </div>
                </div>

                <div className="text-right">
                  <div className="text-sm font-extrabold text-[#1E50A0]">{t.total_xp.toLocaleString()} XP</div>
                  <div className="text-[10px] text-slate-400">{t.total_confirmed} вкладов</div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
