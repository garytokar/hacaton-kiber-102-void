import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { Achievement } from '../types';
import { Trophy, Award, Lock, CheckCircle2, Sparkles, Coins, Flame, BookOpen, Shield, Crown } from 'lucide-react';

export const AchievementsPage: React.FC = () => {
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getAchievements()
      .then(setAchievements)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const unlockedCount = achievements.filter((a) => a.unlocked).length;

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16 select-none">
      {/* Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border border-slate-700/50">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
            <Trophy className="w-3.5 h-3.5 text-amber-400" />
            <span>Награды и признание</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
            Достижения Хранителя
          </h1>
          <p className="text-slate-300 text-xs sm:text-sm mt-1">
            Открывайте уникальные знаки отличия за помощь в сохранении языка.
          </p>
        </div>

        <div className="bg-white/10 backdrop-blur px-5 py-3 rounded-2xl border border-white/20 text-center shrink-0">
          <div className="text-2xl font-black text-amber-400">
            {unlockedCount} / {achievements.length}
          </div>
          <div className="text-[10px] text-slate-300 uppercase tracking-wider font-semibold">
            Разблокировано
          </div>
        </div>
      </div>

      {/* Featured Milestone Card (from Figma) */}
      <div className="bg-white rounded-3xl border border-slate-200/90 p-6 shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="w-1 h-4 bg-[#1E50A0] rounded-full inline-block" />
              <h3 className="font-bold text-base text-[#0C2340]">Семидневный хранитель</h3>
              <span className="text-[10px] font-bold text-[#1E50A0] bg-[#EEF2FF] px-2.5 py-0.5 rounded-full">
                5 / 7 дней
              </span>
            </div>
            <p className="text-xs text-slate-500">
              Подтверждайте переводы 7 дней подряд без перерыва. Награда: +100 XP и редкий бейдж «Огонь эпоса».
            </p>
          </div>

          <div className="w-full sm:w-60 shrink-0 space-y-1.5">
            <div className="flex justify-between text-[11px] font-semibold text-slate-500">
              <span>Прогресс серии</span>
              <span className="text-[#1E50A0] font-bold">71%</span>
            </div>
            <div className="w-full h-2.5 bg-slate-100 rounded-full overflow-hidden">
              <div
                className="h-full bg-[#1E50A0] rounded-full transition-all duration-500"
                style={{ width: '71%' }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Grid */}
      {loading ? (
        <div className="text-center py-12 text-slate-400 text-sm">
          Загрузка наград...
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {achievements.map((ach) => (
            <div
              key={ach.id}
              className={`rounded-3xl p-5 border transition-all flex flex-col justify-between space-y-4 ${
                ach.unlocked
                  ? 'bg-white border-blue-200 shadow-sm'
                  : 'bg-slate-50/70 border-slate-200 opacity-60'
              }`}
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  {/* Circular badge */}
                  <div
                    className={`w-12 h-12 rounded-full flex items-center justify-center text-white shadow-sm ${
                      ach.unlocked
                        ? 'bg-[#0C2340]'
                        : 'bg-slate-300'
                    }`}
                  >
                    {ach.unlocked ? (
                      <Trophy className="w-6 h-6 text-amber-400" />
                    ) : (
                      <Lock className="w-5 h-5 text-slate-500" />
                    )}
                  </div>

                  {ach.unlocked ? (
                    <span className="flex items-center gap-1 text-[11px] font-bold text-[#1E50A0] bg-[#EEF2FF] px-2.5 py-0.5 rounded-full">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      <span>Получено</span>
                    </span>
                  ) : (
                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                      Заблокировано
                    </span>
                  )}
                </div>

                <div>
                  <h3 className="font-extrabold text-base text-[#0C2340]">
                    {ach.name}
                  </h3>
                  <p className="text-xs text-slate-500 mt-1 leading-relaxed">
                    {ach.description}
                  </p>
                </div>
              </div>

              {/* Rewards footer */}
              <div className="pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
                <span className="text-slate-400 font-medium">Награда:</span>
                <div className="flex items-center gap-2 font-bold">
                  <span className="text-[#1E50A0]">+{ach.reward_xp} XP</span>
                  <span className="text-amber-600">+{ach.reward_currency} Ҡурай</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
