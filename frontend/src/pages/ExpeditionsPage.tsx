import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { WeeklyExpedition } from '../types';
import { Compass, Sparkles, Coins, Users, CheckCircle2, ArrowRight } from 'lucide-react';
import { Link } from 'react-router-dom';

export const ExpeditionsPage: React.FC = () => {
  const [expedition, setExpedition] = useState<WeeklyExpedition | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getWeeklyExpedition()
      .then(setExpedition)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16 select-none">
      {/* Banner */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md border border-slate-700/50">
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
          <Compass className="w-3.5 h-3.5 text-amber-400" />
          <span>Коллективные походы за знаниями</span>
        </div>
        <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
          Недельные языковые экспедиции
        </h1>
        <p className="text-slate-300 text-xs sm:text-sm mt-1">
          Объединяйтесь со всеми школами республики для выполнения глобальных миссий.
        </p>
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-400 text-sm">
          Загрузка экспедиции...
        </div>
      ) : expedition ? (
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
              <span>Совместный прогресс школ</span>
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
              <span>Внести свой вклад (+5 XP)</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-3xl p-12 text-center text-slate-400">
          Новая экспедиция готовится к запуску!
        </div>
      )}
    </div>
  );
};
