import React, { useEffect } from 'react';
import confetti from 'canvas-confetti';
import { TaskSubmissionResult } from '../../types';
import { Sparkles, Trophy, Flame, Coins, Check, ArrowRight } from 'lucide-react';

interface RewardModalProps {
  result: TaskSubmissionResult | null;
  onClose: () => void;
  onNextTask?: () => void;
}

export const RewardModal: React.FC<RewardModalProps> = ({ result, onClose, onNextTask }) => {
  useEffect(() => {
    if (result?.success) {
      confetti({
        particleCount: 80,
        spread: 60,
        origin: { y: 0.6 },
        colors: ['#0C2340', '#1E50A0', '#F59E0B', '#3B82F6'],
      });
    }
  }, [result]);

  if (!result) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-xs animate-fade-in select-none">
      <div className="bg-white rounded-3xl max-w-md w-full p-6 text-center shadow-2xl border border-slate-200 transform scale-100 transition-transform">
        {/* Top Icon */}
        <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-[#0C2340] flex items-center justify-center text-white shadow-md">
          {result.level_up ? (
            <Trophy className="w-8 h-8 text-amber-400" />
          ) : (
            <Sparkles className="w-8 h-8 text-amber-400" />
          )}
        </div>

        {/* Header */}
        <h3 className="text-xl font-extrabold text-[#0C2340] mb-1">
          {result.level_up ? '🎊 Новый уровень Хранителя!' : 'Отличный вклад в язык!'}
        </h3>
        <p className="text-xs text-slate-500 mb-6">{result.message}</p>

        {/* Badges Earned */}
        <div className="grid grid-cols-2 gap-3 mb-6">
          <div className="bg-[#EEF2FF] border border-[#1E50A0]/20 p-3 rounded-2xl flex items-center justify-center gap-2">
            <Sparkles className="w-5 h-5 text-[#1E50A0]" />
            <div className="text-left">
              <span className="block text-xs text-slate-500 font-medium">Опыт</span>
              <span className="text-base font-extrabold text-[#1E50A0]">+{result.xp_awarded} XP</span>
            </div>
          </div>

          <div className="bg-[#FFFBEB] border border-amber-200 p-3 rounded-2xl flex items-center justify-center gap-2">
            <Coins className="w-5 h-5 text-amber-500 fill-amber-500" />
            <div className="text-left">
              <span className="block text-xs text-slate-500 font-medium">Валюта</span>
              <span className="text-base font-extrabold text-amber-600">+{result.currency_awarded} Ҡурай</span>
            </div>
          </div>
        </div>

        {/* Streak notification */}
        {result.streak_updated && (
          <div className="bg-amber-50 border border-amber-200 p-3 rounded-2xl mb-4 flex items-center justify-center gap-2 text-amber-800 text-xs font-semibold">
            <Flame className="w-4 h-4 text-amber-500 fill-amber-500" />
            <span>Серия активности увеличена: {result.current_streak} дн.!</span>
          </div>
        )}

        {/* Achievement unlocked if any */}
        {result.achievement_unlocked && (
          <div className="bg-gradient-to-r from-amber-500/10 to-blue-500/10 border border-amber-300 p-3.5 rounded-2xl mb-4 text-left">
            <div className="flex items-center gap-2 mb-1">
              <Trophy className="w-4 h-4 text-amber-500" />
              <span className="text-xs font-bold text-[#0C2340]">Разблокировано достижение!</span>
            </div>
            <div className="text-sm font-extrabold text-[#1E50A0]">
              {result.achievement_unlocked.name}
            </div>
            <div className="text-xs text-slate-500">
              {result.achievement_unlocked.description}
            </div>
          </div>
        )}

        {/* Action buttons */}
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 py-3 px-4 rounded-full border border-slate-200 text-xs font-bold text-slate-700 hover:bg-slate-50 transition-colors"
          >
            Закрыть
          </button>
          {onNextTask && (
            <button
              onClick={() => {
                onClose();
                onNextTask();
              }}
              className="flex-1 py-3 px-4 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold shadow transition-all flex items-center justify-center gap-1.5"
            >
              <span>Дальше</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
