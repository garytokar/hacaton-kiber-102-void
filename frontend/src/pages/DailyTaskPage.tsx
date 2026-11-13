import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Fragment, Translation, TaskSubmissionResult } from '../types';
import { RewardModal } from '../components/gamification/RewardModal';
import instrumentTask from '../assets/figma/instrument_task.png';
import {
  Volume2,
  CheckCircle2,
  Check,
  Edit3,
  Bookmark,
  Sparkles,
  ChevronRight,
  ChevronLeft,
  History,
  ThumbsUp,
  ArrowRight,
  BookOpen,
  AlertCircle
} from 'lucide-react';

export const DailyTaskPage: React.FC = () => {
  const { user, refreshUser } = useAuth();
  const [fragment, setFragment] = useState<Fragment | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [isProposing, setIsProposing] = useState(false);
  const [proposalText, setProposalText] = useState('');
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyList, setHistoryList] = useState<any[]>([]);
  const [rewardResult, setRewardResult] = useState<TaskSubmissionResult | null>(null);
  const [dictSaved, setDictSaved] = useState(false);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  const loadDailyFragment = async () => {
    setLoading(true);
    setIsProposing(false);
    setProposalText('');
    setDictSaved(false);
    try {
      const data = await api.getDailyFragment();
      setFragment(data);
    } catch (err: any) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDailyFragment();
  }, []);

  const handleConfirm = async (translationId: number) => {
    if (!user) {
      alert('Пожалуйста, войдите в аккаунт для сохранения результата и получения XP!');
      return;
    }
    setSubmitting(true);
    try {
      const res = await api.voteTranslation(translationId, 'confirm');
      setRewardResult(res);
      await refreshUser();
      if (fragment) {
        const updated = await api.getFragmentById(fragment.id);
        setFragment(updated);
      }
    } catch (err: any) {
      alert(err.message || 'Ошибка голосования');
    } finally {
      setSubmitting(false);
    }
  };

  const handlePropose = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) {
      alert('Пожалуйста, войдите в аккаунт для предложения перевода!');
      return;
    }
    if (!fragment || !proposalText.trim()) return;

    setSubmitting(true);
    try {
      const res = await api.proposeTranslation({
        fragment_id: fragment.id,
        translation_text: proposalText.trim(),
      });
      setRewardResult(res);
      setIsProposing(false);
      setProposalText('');
      await refreshUser();
      const updated = await api.getFragmentById(fragment.id);
      setFragment(updated);
    } catch (err: any) {
      alert(err.message || 'Ошибка отправки перевода');
    } finally {
      setSubmitting(false);
    }
  };

  const handleSaveToDictionary = async () => {
    if (!fragment || !user) {
      alert('Пожалуйста, войдите в аккаунт, чтобы сохранять слова в личный словарь!');
      return;
    }
    try {
      const topTrans = fragment.translations[0]?.translation_text || 'Перевод фрагмента';
      await api.addDictionaryWord({
        fragment_id: fragment.id,
        word: fragment.text_ba,
        translation: topTrans,
        context: fragment.source_title,
        difficulty: fragment.difficulty,
      });
      setDictSaved(true);
    } catch (err: any) {
      alert(err.message || 'Слово уже есть в словаре');
    }
  };

  const handlePlayAudio = () => {
    if (!fragment) return;
    setIsPlayingAudio(true);
    if ('speechSynthesis' in window) {
      const utterance = new SpeechSynthesisUtterance(fragment.text_ba);
      utterance.lang = 'ru-RU';
      utterance.onend = () => setIsPlayingAudio(false);
      utterance.onerror = () => setIsPlayingAudio(false);
      window.speechSynthesis.speak(utterance);
    } else {
      setTimeout(() => setIsPlayingAudio(false), 1500);
    }
  };

  const loadHistory = async (translationId: number) => {
    try {
      const hist = await api.getTranslationHistory(translationId);
      setHistoryList(hist);
      setHistoryOpen(true);
    } catch (err) {
      console.error(err);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[50vh] flex items-center justify-center">
        <div className="flex items-center gap-2 text-slate-500 font-semibold text-sm">
          <Sparkles className="w-5 h-5 animate-spin text-[#1E50A0]" />
          <span>Загрузка башкирского фрагмента...</span>
        </div>
      </div>
    );
  }

  if (!fragment) {
    return (
      <div className="bg-white rounded-3xl p-8 text-center max-w-md mx-auto shadow-sm border border-slate-200">
        <h3 className="font-bold text-lg text-[#0C2340]">Все фрагменты проверены!</h3>
        <p className="text-xs text-slate-500 mt-2">
          Вы великолепно потрудились сегодня на благо сохранения языка.
        </p>
      </div>
    );
  }

  const primaryTranslation = fragment.translations[0];
  const alternativeTranslations = fragment.translations.slice(1);

  return (
    <div className="max-w-5xl mx-auto space-y-6 pb-16 select-none">
      {/* Top Breadcrumb & Actions Bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-500">
          <span>Обучение</span>
          <ChevronRight className="w-3.5 h-3.5" />
          <span className="text-[#0C2340] font-bold">Ежедневное задание</span>
        </div>

        <button
          onClick={loadDailyFragment}
          className="text-xs font-bold text-[#1E50A0] hover:text-[#0C2340] flex items-center gap-1.5 transition-colors"
        >
          <span>Следующий фрагмент</span>
          <ArrowRight className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Main Task Card (Figma Style) */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 lg:p-10 space-y-6">
        {/* Header Tags & Bookmark */}
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 pb-4">
          <div className="flex items-center gap-2">
            <span className="px-3.5 py-1 rounded-full text-xs font-bold bg-[#EEF2FF] text-[#1E50A0]">
              {fragment.category}
            </span>
            <span className="px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider bg-slate-100 text-slate-600">
              {fragment.difficulty === 'easy' ? 'Базовый' : fragment.difficulty === 'hard' ? 'Сложный (+8 XP)' : 'Средний'}
            </span>
          </div>

          <button
            onClick={handleSaveToDictionary}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold transition-colors ${
              dictSaved
                ? 'bg-blue-50 text-[#1E50A0]'
                : 'bg-slate-50 hover:bg-slate-100 text-slate-700 border border-slate-200'
            }`}
          >
            <Bookmark className={`w-3.5 h-3.5 ${dictSaved ? 'fill-[#1E50A0]' : ''}`} />
            <span>{dictSaved ? 'Сохранено в словарь' : 'В мой словарь'}</span>
          </button>
        </div>

        {/* Content Split: Left Text + Right Musical Instrument Overlay */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center">
          {/* Left Text */}
          <div className="lg:col-span-7 space-y-4">
            <div className="space-y-2">
              <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">
                Оригинал на башкирском языке
              </span>
              <div className="text-2xl sm:text-3xl font-black text-[#0C2340] leading-snug tracking-tight font-serif">
                «{fragment.text_ba}»
              </div>
              <div className="flex items-center gap-2 text-xs text-slate-500 pt-1">
                <BookOpen className="w-3.5 h-3.5 text-[#1E50A0]" />
                <span className="font-semibold">{fragment.source_title}</span>
                {fragment.source_document && (
                  <span className="text-slate-400">• {fragment.source_document}</span>
                )}
              </div>
            </div>

            {/* Proposed Translation Block */}
            {primaryTranslation ? (
              <div className="bg-[#F8FAFC] rounded-2xl p-5 border border-slate-200/80 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                    Предложенный литературный перевод
                  </span>
                  <span className="text-xs font-bold text-[#1E50A0] bg-[#EEF2FF] px-2.5 py-0.5 rounded-full">
                    Голосов: {primaryTranslation.votes_count}
                  </span>
                </div>

                <div className="text-base sm:text-lg font-bold text-slate-800">
                  «{primaryTranslation.translation_text}»
                </div>

                <div className="flex items-center justify-between text-[11px] text-slate-400 pt-2 border-t border-slate-200/50">
                  <span>Добавил: {primaryTranslation.creator_username || 'Система'}</span>
                  <button
                    onClick={() => loadHistory(primaryTranslation.id)}
                    className="hover:text-[#1E50A0] flex items-center gap-1 font-semibold"
                  >
                    <History className="w-3.5 h-3.5" />
                    <span>История правок</span>
                  </button>
                </div>
              </div>
            ) : (
              <div className="bg-amber-50 rounded-2xl p-5 border border-amber-200 text-amber-800 text-xs">
                К этому фрагменту пока нет предложенного перевода. Станьте первым!
              </div>
            )}
          </div>

          {/* Right Instrument Audio Box */}
          <div className="lg:col-span-5 flex justify-center">
            <div className="relative w-full max-w-sm rounded-2xl overflow-hidden shadow-sm group border border-slate-100">
              <img
                src={instrumentTask}
                alt="Башкирский инструмент"
                className="w-full h-52 sm:h-56 object-cover transition-transform group-hover:scale-105 duration-300"
              />
              <button
                onClick={handlePlayAudio}
                disabled={isPlayingAudio}
                className={`absolute inset-0 m-auto w-12 h-12 rounded-full bg-white/95 text-[#0C2340] shadow-xl flex items-center justify-center transition-all ${
                  isPlayingAudio ? 'scale-110 ring-4 ring-[#1E50A0]/40' : 'hover:scale-110 hover:bg-white'
                }`}
                title="Озвучить башкирский фрагмент"
              >
                <Volume2 className={`w-5 h-5 ${isPlayingAudio ? 'text-[#1E50A0] animate-pulse' : 'text-[#0C2340]'}`} />
              </button>
            </div>
          </div>
        </div>

        {/* Action Buttons: Confirm or Propose */}
        <div className="space-y-4 pt-4 border-t border-slate-100">
          <div className="flex flex-col sm:flex-row gap-3">
            {primaryTranslation && (
              <button
                onClick={() => handleConfirm(primaryTranslation.id)}
                disabled={submitting || primaryTranslation.user_voted === 'confirm'}
                className="flex-1 py-3.5 px-6 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white font-bold text-xs sm:text-sm shadow-md flex items-center justify-center gap-2 transition-all disabled:opacity-50"
              >
                <Check className="w-4 h-4 text-white" />
                <span>
                  {primaryTranslation.user_voted === 'confirm'
                    ? 'Вы подтвердили этот перевод ✓'
                    : 'Подтвердить перевод (+5 XP)'}
                </span>
              </button>
            )}

            <button
              onClick={() => setIsProposing(!isProposing)}
              className="flex-1 py-3.5 px-6 rounded-full bg-white hover:bg-blue-50 border border-[#1E50A0] text-[#1E50A0] font-bold text-xs sm:text-sm flex items-center justify-center gap-2 transition-colors"
            >
              <Edit3 className="w-4 h-4 text-[#1E50A0]" />
              <span>Предложить исправление (+8 XP)</span>
            </button>
          </div>

          {/* Proposal Input Drawer */}
          {isProposing && (
            <form onSubmit={handlePropose} className="bg-slate-50 rounded-2xl p-5 border border-slate-200 space-y-3 animate-fade-in">
              <label className="block text-xs font-bold text-slate-700">
                Ваш вариант литературного перевода на русский язык:
              </label>
              <textarea
                required
                rows={3}
                value={proposalText}
                onChange={(e) => setProposalText(e.target.value)}
                placeholder="Введите точный перевод фразы..."
                className="w-full p-3 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-[#1E50A0] text-xs sm:text-sm bg-white"
              />
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsProposing(false)}
                  className="px-4 py-2 rounded-full text-xs font-semibold text-slate-500 hover:bg-slate-200"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-5 py-2 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold shadow transition-all"
                >
                  Отправить на голосование
                </button>
              </div>
            </form>
          )}
        </div>

        {/* Alternative Translations List */}
        {alternativeTranslations.length > 0 && (
          <div className="space-y-3 pt-4 border-t border-slate-100">
            <h4 className="text-xs font-bold text-slate-500 uppercase tracking-wider">
              Другие предложенные переводы сообщества ({alternativeTranslations.length})
            </h4>
            <div className="space-y-2">
              {alternativeTranslations.map((alt) => (
                <div
                  key={alt.id}
                  className="p-3.5 rounded-xl border border-slate-200 bg-white flex items-center justify-between gap-4"
                >
                  <div>
                    <div className="text-xs font-semibold text-slate-900">
                      «{alt.translation_text}»
                    </div>
                    <div className="text-[10px] text-slate-400 mt-0.5">
                      Автор: {alt.creator_username || 'Хранитель'} • Голосов: {alt.votes_count}
                    </div>
                  </div>

                  <button
                    onClick={() => handleConfirm(alt.id)}
                    disabled={submitting || alt.user_voted === 'confirm'}
                    className={`px-3 py-1.5 rounded-full text-xs font-bold flex items-center gap-1 transition-colors ${
                      alt.user_voted === 'confirm'
                        ? 'bg-blue-100 text-[#1E50A0]'
                        : 'bg-slate-100 hover:bg-blue-50 text-slate-700 hover:text-[#1E50A0]'
                    }`}
                  >
                    <ThumbsUp className="w-3.5 h-3.5" />
                    <span>{alt.votes_count}</span>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* History Modal */}
      {historyOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-3xl max-w-lg w-full p-6 space-y-4 shadow-2xl border border-slate-200 animate-scale-up">
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-base text-[#0C2340]">История изменений перевода</h3>
              <button
                onClick={() => setHistoryOpen(false)}
                className="text-slate-400 hover:text-slate-600 text-lg font-bold"
              >
                ✕
              </button>
            </div>

            <div className="space-y-2 max-h-80 overflow-y-auto pr-1">
              {historyList.length === 0 ? (
                <p className="text-xs text-slate-400 py-4 text-center">История правок пуста</p>
              ) : (
                historyList.map((h, i) => (
                  <div key={i} className="p-3 rounded-xl bg-slate-50 text-xs border border-slate-100 space-y-1">
                    <div className="flex justify-between text-slate-400 text-[10px]">
                      <span>{h.user_name || 'Пользователь'}</span>
                      <span>{new Date(h.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                    <div className="text-slate-800 font-semibold">{h.old_text ? `«${h.old_text}» → ` : ''}«{h.new_text}»</div>
                    {h.comment && <div className="text-slate-500 text-[11px] italic">{h.comment}</div>}
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}

      {/* Gamification Reward Modal */}
      {rewardResult && (
        <RewardModal
          result={rewardResult}
          onClose={() => setRewardResult(null)}
        />
      )}
    </div>
  );
};
