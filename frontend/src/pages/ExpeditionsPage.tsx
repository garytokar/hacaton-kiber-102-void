import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { WeeklyExpedition } from '../types';
import { UserAvatar } from '../components/ui/UserAvatar';
import {
  Compass,
  Sparkles,
  Coins,
  Users,
  CheckCircle2,
  ArrowRight,
  School,
  Trophy,
  Award,
  Shield,
  Layers,
  Flag
} from 'lucide-react';
import { Link } from 'react-router-dom';

export const ExpeditionsPage: React.FC = () => {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'team' | 'expedition'>('team');
  const [expedition, setExpedition] = useState<WeeklyExpedition | null>(null);
  const [teamData, setTeamData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [expRes, teamRes] = await Promise.all([
        api.getWeeklyExpedition().catch(() => null),
        api.getMyTeam().catch(() => null),
      ]);
      setExpedition(expRes);
      setTeamData(teamRes);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* Header Banner */}
      <div className="yan-page-hero bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md border border-slate-700/50 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <Users className="w-3.5 h-3.5 text-amber-400" />
            <span>Командная геймификация и экспедиции</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Команда школы и Экспедиции
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1 max-w-xl">
            Соревнуйтесь вместе с одноклассниками, выполняйте общие цели школы и участвуйте в глобальных походах за знаниями.
          </p>
        </div>

        {/* Tab switch */}
        <div className="flex items-center gap-1.5 bg-white/10 p-1.5 rounded-2xl border border-white/10 text-xs font-bold shrink-0">
          <button
            onClick={() => setActiveTab('team')}
            className={`px-4 py-2 rounded-xl transition-all ${
              activeTab === 'team' ? 'bg-[#1E50A0] text-white shadow-sm' : 'text-slate-300 hover:text-white'
            }`}
          >
            Моя школа и класс
          </button>
          <button
            onClick={() => setActiveTab('expedition')}
            className={`px-4 py-2 rounded-xl transition-all ${
              activeTab === 'expedition' ? 'bg-[#1E50A0] text-white shadow-sm' : 'text-slate-300 hover:text-white'
            }`}
          >
            Экспедиция недели
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-16 text-slate-400 text-sm">
          Загрузка командных данных...
        </div>
      ) : activeTab === 'team' ? (
        /* TAB 1: SCHOOL TEAM */
        <div className="space-y-6">
          {teamData?.has_school ? (
            <>
              {/* School Hero Card */}
              <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 space-y-6">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 pb-5">
                  <div className="flex items-center gap-4">
                    <div className="w-14 h-14 rounded-2xl bg-[#EAF3FF] text-[#5E9BE0] flex items-center justify-center shrink-0">
                      <School className="w-7 h-7 text-[#5E9BE0]" />
                    </div>
                    <div>
                      <span className="text-[10px] font-bold px-2.5 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] uppercase tracking-wider">
                        Школьная сборная
                      </span>
                      <h2 className="text-xl sm:text-2xl font-black text-[#0C2340] mt-1">
                        {teamData.school.name}
                      </h2>
                      <div className="text-xs text-slate-500 mt-0.5">
                        г. {teamData.school.city} • Ваш класс:{' '}
                        <strong className="text-[#1E50A0]">{teamData.school.class_name || 'Сборная'}</strong>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="text-right">
                      <div className="text-xs font-bold text-slate-400 uppercase">Опыт команды</div>
                      <div className="text-xl font-black text-[#0C2340]">
                        {teamData.school.total_xp} XP
                      </div>
                    </div>
                    <div className="w-10 h-10 rounded-2xl bg-amber-50 border border-amber-200 flex items-center justify-center text-amber-600 font-black text-sm">
                      #{teamData.school.rank}
                    </div>
                  </div>
                </div>

                {/* Collective Mission */}
                <div className="p-5 rounded-2xl bg-[#F8FAFC] border border-slate-200/80 space-y-3">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-2">
                        <Flag className="w-4 h-4 text-emerald-600" />
                        <h4 className="font-extrabold text-sm text-[#0C2340]">
                          {teamData.collective_goal.title}
                        </h4>
                      </div>
                      <p className="text-xs text-slate-500 mt-0.5">
                        Вклад каждого ученика школы приближает общую победу и награду.
                      </p>
                    </div>

                    <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-50 text-emerald-800 font-bold text-xs border border-emerald-200 shrink-0">
                      <span>Награда: +{teamData.collective_goal.reward_xp} XP • +{teamData.collective_goal.reward_currency} Ҡурай</span>
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <div className="flex justify-between text-xs font-bold text-slate-600">
                      <span>Прогресс школы</span>
                      <span className="text-[#1E50A0]">
                        {teamData.collective_goal.current} из {teamData.collective_goal.target} переводов ({teamData.collective_goal.percentage}%)
                      </span>
                    </div>
                    <div className="w-full h-3 bg-slate-200 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-[#0C2340] to-[#1E50A0] rounded-full transition-all duration-500"
                        style={{ width: `${teamData.collective_goal.percentage}%` }}
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Teammates Roster */}
              <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="font-black text-base text-[#0C2340]">
                    Состав команды ({teamData.members?.length || 0} уч.)
                  </h3>
                  <Link
                    to="/leaderboard"
                    className="text-xs font-bold text-[#1E50A0] hover:underline flex items-center gap-1"
                  >
                    <span>Таблица всех школ</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </Link>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                  {teamData.members?.map((m: any, idx: number) => (
                    <div
                      key={m.id}
                      className={`p-3.5 rounded-2xl border transition-all flex items-center gap-3 ${
                        m.is_me
                          ? 'bg-blue-50/70 border-[#1E50A0]/60 ring-1 ring-[#1E50A0]/20'
                          : 'bg-[#F8FAFC] border-slate-200/70'
                      }`}
                    >
                      <UserAvatar
                        user={{
                          ...m,
                          cosmetics: m.cosmetics,
                        }}
                        size="md"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1.5">
                          <span className="font-extrabold text-xs text-[#0C2340] truncate">
                            {m.full_name}
                          </span>
                          {m.is_me && (
                            <span className="text-[9px] font-black px-1.5 py-0.2 rounded-full bg-[#1E50A0] text-white">
                              ВЫ
                            </span>
                          )}
                        </div>
                        <div className="text-[11px] text-slate-400">
                          Уровень {m.level} • <strong className="text-slate-700">{m.xp} XP</strong>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </>
          ) : (
            <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-10 text-center space-y-4 max-w-lg mx-auto">
              <div className="w-16 h-16 rounded-3xl bg-blue-50 text-[#1E50A0] flex items-center justify-center mx-auto">
                <School className="w-8 h-8" />
              </div>
              <div className="space-y-1">
                <h3 className="text-lg font-black text-[#0C2340]">Вы пока без команды</h3>
                <p className="text-xs text-slate-500">
                  Выберите свою школу или лицей в профиле, чтобы объединиться с одноклассниками и поднимать школу в рейтинге Республики.
                </p>
              </div>
              <Link
                to="/profile"
                className="inline-flex items-center gap-2 px-6 py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs shadow-md transition-all"
              >
                <span>Выбрать школу в профиле</span>
                <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
          )}
        </div>
      ) : (
        /* TAB 2: EXPEDITION */
        expedition && (
          <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 pb-5">
              <div>
                <span className="text-[10px] font-bold px-2.5 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] uppercase tracking-wider">
                  Текущая цель недели
                </span>
                <h2 className="text-2xl font-black text-[#0C2340] mt-2">
                  {expedition.title}
                </h2>
                <p className="text-sm text-slate-600 mt-1 max-w-2xl leading-relaxed">
                  {expedition.description}
                </p>
              </div>

              <div className="flex sm:flex-col items-center sm:items-end justify-between sm:justify-center gap-1 bg-[#FFFBEB] p-4 rounded-2xl border border-amber-200 shrink-0">
                <span className="text-xs font-semibold text-slate-500">Награда экспедиции:</span>
                <span className="text-sm sm:text-base font-extrabold text-[#0C2340]">
                  +{expedition.reward_xp} XP • +{expedition.reward_currency} Ҡурай
                </span>
              </div>
            </div>

            {/* Progress bar */}
            <div className="space-y-2">
              <div className="flex justify-between text-xs font-bold text-slate-700">
                <span>Совместный прогресс школ Башкортостана</span>
                <span className="text-[#1E50A0]">
                  {expedition.current_count} из {expedition.target_count} ({expedition.progress_percentage}%)
                </span>
              </div>
              <div className="w-full h-3.5 bg-slate-100 rounded-full overflow-hidden p-0.5">
                <div
                  className="h-full bg-gradient-to-r from-[#0C2340] to-[#1E50A0] rounded-full transition-all duration-500 shadow-sm"
                  style={{ width: `${expedition.progress_percentage}%` }}
                />
              </div>
            </div>

            <div className="pt-2 flex justify-end">
              <Link
                to="/task"
                className="px-6 py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md flex items-center gap-2 transition-all"
              >
                <span>Принять участие в экспедиции</span>
                <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
          </div>
        )
      )}
    </div>
  );
};
