import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Fragment, TaskSubmissionResult, DailyQuest } from '../types';
import { RewardModal } from '../components/gamification/RewardModal';
import { AudioSpeakerButton } from '../components/ui/AudioSpeakerButton';
import { speakBashkir } from '../utils/speech';
import instrumentTask from '../assets/figma/instrument_task.png';
import {
  Volume2,
  CheckCircle2,
  Check,
  Edit3,
  Bookmark,
  Sparkles,
  ChevronRight,
  ThumbsUp,
  ArrowRight,
  BookOpen,
  AlertCircle
} from 'lucide-react';

export const DailyTaskPage: React.FC = () => {
  const { user, refreshUser, openAuthModal } = useAuth();
  const navigate = useNavigate();

  const [fragment, setFragment] = useState<Fragment | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [isProposing, setIsProposing] = useState(false);
  const [proposalText, setProposalText] = useState('');
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyList, setHistoryList] = useState<any[]>([]);
  const [rewardResult, setRewardResult] = useState<TaskSubmissionResult | null>(null);
  const [dictSaved, setDictSaved] = useState(false);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);
  const [dailyQuests, setDailyQuests] = useState<DailyQuest[]>([]);

  // Toast notifications
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'info' | 'error' } | null>(null);

  const showToast = (message: string, type: 'success' | 'info' | 'error' = 'info') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  };

  const loadDailyFragment = async () => {
    setLoading(true);
    setLoadError(null);
    setIsProposing(false);
    setProposalText('');
    setDictSaved(false);
    try {
      const data = await api.getDailyFragment();
      setFragment(data);
      api.getDailyQuests().then(setDailyQuests).catch(() => {});
    } catch (err: any) {
      setLoadError(err.message || 'Не удалось загрузить башкирский фрагмент');
      setFragment(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDailyFragment();
  }, []);

  const handleConfirm = async (translationId: number) => {
    if (!user) {
      showToast('Пожалуйста, войдите в аккаунт для сохранения результата!', 'info');
      openAuthModal('/task');
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
      showToast('Перевод успешно подтверждён! ✓', 'success');
    } catch (err: any) {
      showToast(err.message || 'Ошибка голосования', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handlePropose = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) {
      showToast('Пожалуйста, войдите в аккаунт для предложения перевода!', 'info');
      openAuthModal('/task');
      return;
    }
    if (!fragment || !proposalText.trim()) return;

    setSubmitting(true);
    try {
      const primaryTrans = fragment.translations && fragment.translations.length > 0 ? fragment.translations[0] : null;
      const res = await api.proposeTranslation({
        fragment_id: fragment.id,
        fragment_translation_id: primaryTrans ? (primaryTrans.id as any) : undefined,
        translation_text: proposalText.trim(),
      });
      setRewardResult(res);
      setIsProposing(false);
      setProposalText('');
      await refreshUser();
      const updated = await api.getFragmentById(fragment.id);
      setFragment(updated);
      showToast('Перевод отправлен на голосование! 🎉', 'success');
    } catch (err: any) {
      showToast(err.message || 'Ошибка отправки перевода', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleSaveToDictionary = async () => {
    if (!fragment || !user) {
      showToast('Пожалуйста, войдите в аккаунт, чтобы сохранять слова в личный словарь!', 'info');
      openAuthModal('/task');
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
      showToast('Слово сохранено в личный словарь! ✓', 'success');
    } catch (err: any) {
      showToast(err.message || 'Слово уже есть в словаре', 'error');
    }
  };

  const handlePlayAudio = () => {
    if (!fragment) return;
    setIsPlayingAudio(true);
    speakBashkir(fragment.text_ba, {
      onStart: () => setIsPlayingAudio(true),
      onEnd: () => setIsPlayingAudio(false),
      onError: () => setIsPlayingAudio(false),
    });
  };

  const loadHistory = async (translationId: number) => {
    try {
      const hist = await api.getTranslationHistory(translationId);
      setHistoryList(hist);
      setHistoryOpen(true);
    } catch {
      showToast('Не удалось загрузить историю правок', 'error');
    }
  };

  if (loading) {
    return (
      <div className="bg-white rounded-3xl p-12 text-center max-w-md mx-auto shadow-sm border border-slate-200">
        <div className="flex items-center justify-center gap-2 text-slate-500 font-semibold text-sm">
          <Sparkles className="w-5 h-5 animate-spin text-[#1E50A0]" />
          <span>Загрузка башкирского фрагмента...</span>
        </div>
      </div>
    );
  }

  // Network / Loading error state with retry button
  if (loadError) {
    return (
      <div className="bg-white rounded-3xl p-8 text-center max-w-md mx-auto shadow-sm border border-slate-200 space-y-4">
        <div className="w-12 h-12 rounded-2xl bg-red-50 text-red-500 mx-auto flex items-center justify-center">
          <AlertCircle className="w-6 h-6" />
        </div>
        <h3 className="font-bold text-lg text-[#0C2340]">Ошибка загрузки задания</h3>
        <p className="text-xs text-slate-500 leading-relaxed">{loadError}</p>
        <button
          type="button"
          onClick={loadDailyFragment}
          className="px-5 py-2.5 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white text-xs font-bold transition-colors shadow-xs cursor-pointer"
        >
          Повторить попытку
        </button>
      </div>
    );
  }

  // All fragments completed state with link to Treasury
  if (!fragment) {
    return (
      <div className="bg-white rounded-3xl p-8 text-center max-w-md mx-auto shadow-sm border border-slate-200 space-y-4">
        <div className="w-12 h-12 rounded-2xl bg-emerald-50 text-emerald-600 mx-auto flex items-center justify-center">
          <CheckCircle2 className="w-6 h-6" />
        </div>
        <h3 className="font-bold text-lg text-[#0C2340]">Все фрагменты проверены!</h3>
        <p className="text-xs text-slate-500 leading-relaxed">
          Вы великолепно потрудились сегодня на благо сохранения языка.
        </p>
        <Link
          to="/treasury"
          className="inline-flex items-center gap-2 px-5 py-2.5 rounded-full bg-[#0C2340] hover:bg-[#1E50A0] text-white text-xs font-bold transition-colors shadow-xs"
        >
          <span>Перейти в сокровищницу</span>
          <ArrowRight className="w-3.5 h-3.5" />
        </Link>
      </div>
    );
  }

  const primaryTranslation = fragment.translations[0];
  const alternativeTranslations = fragment.translations.slice(1);

  return (
    <div className="max-w-5xl mx-auto space-y-6 pb-16">
      {/* Toast Notification */}
      {toast && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-fade-in shadow-2xl rounded-2xl px-5 py-3 flex items-center gap-2.5 text-xs font-bold bg-[#0C2340] text-white border border-slate-700">
          {toast.type === 'success' && <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />}
          {toast.type === 'error' && <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />}
          {toast.type === 'info' && <Sparkles className="w-4 h-4 text-blue-400 shrink-0" />}
          <span>{toast.message}</span>
        </div>
      )}

      {/* Top Breadcrumb & Actions Bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-500">
          <span>Обучение</span>
          <ChevronRight className="w-3.5 h-3.5" />
          <span className="text-[#0C2340] font-bold">Ежедневное задание</span>
        </div>

        <button
          type="button"
          onClick={loadDailyFragment}
          className="text-xs font-bold text-[#1E50A0] hover:text-[#0C2340] flex items-center gap-1.5 transition-colors cursor-pointer"
        >
          <span>Следующий фрагмент</span>
          <ArrowRight className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Main Task Card */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-8 lg:p-10 space-y-6">
        {/* Header Tags & Bookmark */}
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 pb-4">
          <div className="flex items-center gap-2">
            <span className="px-3.5 py-1 rounded-full text-xs font-bold bg-[#EEF2FF] text-[#1E50A0]">
              {fragment.category}
            </span>
            <span className="px-2.5 py-1 rounded-full text-xs font-semibold uppercase tracking-wider bg-slate-100 text-slate-600">
              {fragment.difficulty === 'easy' ? 'Базовый' : fragment.difficulty === 'hard' ? 'Сложный (+8 XP)' : 'Средний'}
            </span>
          </div>

          <button
            type="button"
            onClick={handleSaveToDictionary}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold transition-colors cursor-pointer ${
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
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
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

              <div className="pt-1">
                <AudioSpeakerButton
                  text={fragment.text_ba}
                  size="sm"
                  showLabel={true}
                  label="Слушать башкирское произношение"
                />
              </div>
            </div>

            {/* Proposed Translation Block */}
            {primaryTranslation ? (
              <div className="bg-[#F8FAFC] rounded-2xl p-5 border border-slate-200/80 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                    Основной перевод
                  </span>
                  <button
                    type="button"
                    onClick={() => loadHistory(primaryTranslation.id)}
                    className="text-xs text-[#1E50A0] hover:underline flex items-center gap-1 font-medium cursor-pointer"
                  >
                    История правок
                  </button>
                </div>

                <div className="text-base sm:text-lg font-bold text-slate-800">
                  {primaryTranslation.translation_text}
                </div>

                <div className="flex items-center justify-between pt-1">
                  <div className="text-xs text-slate-500">
                    Автор: <span className="font-bold text-slate-700">{primaryTranslation.creator_username || 'Хранитель'}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-xs text-emerald-600 font-bold bg-emerald-50 px-2.5 py-0.5 rounded-full">
                    <Check className="w-3 h-3" />
                    <span>{primaryTranslation.votes_count} подтверждений</span>
                  </div>
                </div>

                {/* Consensus & Status Feedback Loop from ТЗ */}
                <div className="pt-2.5 border-t border-slate-200/60 space-y-2">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center gap-1.5">
                      <span
                        className={`px-2.5 py-0.5 rounded-full text-xs font-bold border flex items-center gap-1.5 ${
                          primaryTranslation.status === 'confirmed'
                            ? 'bg-emerald-50 text-emerald-800 border-emerald-300'
                            : primaryTranslation.status === 'expert_review'
                            ? 'bg-amber-50 text-amber-800 border-amber-300'
                            : 'bg-blue-50 text-[#1E50A0] border-blue-200'
                        }`}
                      >
                        <span className={`w-1.5 h-1.5 rounded-full ${
                          primaryTranslation.status === 'confirmed' ? 'bg-emerald-500' : primaryTranslation.status === 'expert_review' ? 'bg-amber-500' : 'bg-blue-500'
                        }`} />
                        <span>
                          {primaryTranslation.status === 'confirmed'
                            ? 'Статус: Подтверждён сообществом'
                            : primaryTranslation.status === 'expert_review'
                            ? 'Статус: На экспертизе РАН'
                            : 'Статус: На проверке (pending)'}
                        </span>
                      </span>
                    </div>

                    <div className="flex items-center gap-1.5">
                      <div className="flex items-center gap-1">
                        {[1, 2, 3].map((step) => (
                          <span
                            key={step}
                            className={`w-2.5 h-2.5 rounded-full transition-all ${
                              primaryTranslation.votes_count >= step
                                ? 'bg-[#1E50A0]'
                                : 'bg-slate-200'
                            }`}
                          />
                        ))}
                      </div>
                      <span className="text-xs font-bold text-[#0C2340]">
                        {Math.min(primaryTranslation.votes_count, 3)} из 3 голосов
                      </span>
                    </div>
                  </div>

                  {primaryTranslation.status === 'confirmed' || primaryTranslation.votes_count >= 3 ? (
                    <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-2.5 text-xs text-emerald-800 font-semibold flex items-center gap-2">
                      <Check className="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                      <span>
                        ✓ Консенсус достигнут! Перевод добавлен в{' '}
                        <Link to="/treasury" className="underline font-bold hover:text-emerald-950">
                          Сокровищницу языка
                        </Link>
                      </span>
                    </div>
                  ) : (
                    <div className="text-xs text-slate-500 flex items-center gap-1">
                      <span>💡</span>
                      <span>
                        Ещё {Math.max(1, 3 - primaryTranslation.votes_count)} голос(а) до переноса в Сокровищницу
                      </span>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="bg-amber-50 rounded-2xl p-4 border border-amber-200 text-amber-800 text-xs">
                Пока нет перевода для этого фрагмента. Будьте первым, кто его предложит!
              </div>
            )}
          </div>

          {/* Right Cultural Art Box */}
          <div className="lg:col-span-5 relative rounded-2xl overflow-hidden h-64 border border-slate-200 shadow-sm group">
            <img
              src={instrumentTask}
              alt="Думбыра"
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent flex flex-col justify-end p-5 text-white">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-300">
                Культурный контекст
              </span>
              <div className="font-bold text-sm leading-tight mt-1">
                «Язык предков звучит в мелодиях думбыры и мудрости пословиц»
              </div>
              <button
                type="button"
                onClick={handlePlayAudio}
                className="mt-3 inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-white/20 hover:bg-white/30 backdrop-blur-md text-xs font-bold transition-colors w-fit cursor-pointer"
              >
                <Volume2 className="w-3.5 h-3.5" />
                <span>{isPlayingAudio ? 'Воспроизведение...' : 'Слушать фразу'}</span>
              </button>
            </div>
          </div>
        </div>

        {/* Action Buttons: Confirm or Propose */}
        <div className="pt-2 border-t border-slate-100 flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            {primaryTranslation && (
              <button
                type="button"
                onClick={() => handleConfirm(primaryTranslation.id)}
                disabled={submitting || primaryTranslation.user_voted === 'confirm'}
                className={`px-6 py-3 rounded-full text-xs font-bold flex items-center gap-2 transition-all shadow-sm cursor-pointer ${
                  primaryTranslation.user_voted === 'confirm'
                    ? 'bg-emerald-600 text-white cursor-default'
                    : 'bg-[#1E50A0] hover:bg-[#2563EB] text-white hover:shadow-md'
                }`}
              >
                <Check className="w-4 h-4" />
                <span>
                  {primaryTranslation.user_voted === 'confirm' ? 'Вы подтвердили этот перевод ✓' : 'Подтвердить перевод'}
                </span>
              </button>
            )}

            <button
              type="button"
              onClick={() => setIsProposing(!isProposing)}
              className="px-5 py-3 rounded-full bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold flex items-center gap-2 transition-colors cursor-pointer"
            >
              <Edit3 className="w-4 h-4" />
              <span>{isProposing ? 'Отмена' : 'Предложить свой перевод'}</span>
            </button>
          </div>
        </div>

        {/* Propose Translation Form */}
        {isProposing && (
          <form onSubmit={handlePropose} className="bg-slate-50 rounded-2xl p-5 border border-slate-200 space-y-3 animate-fade-in">
            <h4 className="text-xs font-bold text-[#0C2340]">Ваш вариант перевода на русский язык:</h4>
            <textarea
              required
              rows={3}
              value={proposalText}
              onChange={(e) => setProposalText(e.target.value)}
              placeholder="Введите точный перевод фразы..."
              className="w-full p-3.5 rounded-xl border border-slate-300 text-xs focus:outline-none focus:ring-2 focus:ring-[#1E50A0] bg-white"
            />
            <div className="flex items-center justify-end gap-2">
              <button
                type="button"
                onClick={() => setIsProposing(false)}
                className="px-4 py-2 rounded-full border border-slate-200 text-xs font-semibold text-slate-600 hover:bg-slate-100 cursor-pointer"
              >
                Отмена
              </button>
              <button
                type="submit"
                disabled={submitting}
                className="px-5 py-2 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs font-bold shadow transition-all cursor-pointer"
              >
                Отправить на голосование
              </button>
            </div>
          </form>
        )}

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
                    <div className="text-xs text-slate-400 mt-0.5">
                      Автор: {alt.creator_username || 'Хранитель'} • Голосов: {alt.votes_count}
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() => handleConfirm(alt.id)}
                    disabled={submitting || alt.user_voted === 'confirm'}
                    className={`px-3 py-1.5 rounded-full text-xs font-bold flex items-center gap-1 transition-colors cursor-pointer ${
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

      {/* Daily Quests Block */}
      {dailyQuests.length > 0 && (
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-sm p-6 sm:p-7 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-slate-100">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-amber-500" />
              <h3 className="text-sm font-bold text-[#0C2340]">Ежедневные квесты на сегодня</h3>
            </div>
            <span className="text-xs text-slate-400 font-semibold">
              {dailyQuests.filter((q) => q.completed).length} из {dailyQuests.length} завершено
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
            {dailyQuests.slice(0, 2).map((quest) => (
              <div
                key={quest.id}
                className={`p-4 rounded-2xl border transition-all ${
                  quest.completed
                    ? 'bg-emerald-50/60 border-emerald-200'
                    : 'bg-slate-50/70 border-slate-200/80'
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold text-[#0C2340]">{quest.name}</span>
                      {quest.completed && (
                        <span className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-100 px-2 py-0.5 rounded-full">
                          <Check className="w-3 h-3" />
                          Выполнен
                        </span>
                      )}
                    </div>
                    <p className="text-[11px] text-slate-500 leading-snug">{quest.description}</p>
                  </div>
                  <div className="shrink-0 text-right">
                    <span className="inline-block px-2 py-0.5 rounded-full text-[11px] font-bold bg-amber-50 text-amber-700 border border-amber-200">
                      +{quest.reward_xp} XP
                    </span>
                  </div>
                </div>

                <div className="mt-3 space-y-1">
                  <div className="flex justify-between text-[11px] font-semibold text-slate-500">
                    <span>Прогресс</span>
                    <span>
                      {Math.min(quest.progress, quest.target)} / {quest.target}
                    </span>
                  </div>
                  <div className="w-full h-1.5 bg-slate-200 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        quest.completed ? 'bg-emerald-500' : 'bg-[#1E50A0]'
                      }`}
                      style={{
                        width: `${Math.min(100, Math.round((quest.progress / quest.target) * 100))}%`,
                      }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* History Modal */}
      {historyOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-3xl max-w-lg w-full p-6 space-y-4 shadow-2xl border border-slate-200 animate-scale-up">
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-base text-[#0C2340]">История изменений перевода</h3>
              <button
                type="button"
                onClick={() => setHistoryOpen(false)}
                className="text-slate-400 hover:text-slate-600 text-lg font-bold cursor-pointer"
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
                    <div className="flex justify-between text-slate-400 text-xs">
                      <span>{h.user_name || 'Пользователь'}</span>
                      <span>{new Date(h.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                    <div className="text-slate-800 font-semibold">{h.old_text ? `«${h.old_text}» → ` : ''}«{h.new_text}»</div>
                    {h.comment && <div className="text-slate-500 text-xs italic">{h.comment}</div>}
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
