import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { api } from '../api/client';
import { DailyQuest, WeeklyExpedition, Achievement, Fragment } from '../types';
import heroMountains from '../assets/figma/hero_mountains.png';
import instrumentTask from '../assets/figma/instrument_task.png';
import treasurySunrise from '../assets/figma/treasury_sunrise.png';
import {
  Sparkles,
  Flame,
  Coins,
  Shield,
  ArrowRight,
  CheckCircle2,
  Trophy,
  Compass,
  BookA,
  Target,
  Calendar,
  Volume2,
  Check,
  Edit3,
  Heart,
  Crown,
  BookOpen,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { user, refreshUser } = useAuth();
  const [quests, setQuests] = useState<DailyQuest[]>([]);
  const [expedition, setExpedition] = useState<WeeklyExpedition | null>(null);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [loading, setLoading] = useState(true);

  // Daily task fragments
  const [fragments, setFragments] = useState<Fragment[]>([]);
  const [currentIdx, setCurrentIdx] = useState(0);
  const [confirmedIds, setConfirmedIds] = useState<number[]>([]);
  const [taskNotice, setTaskNotice] = useState<string | null>(null);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  // Leaderboard tab state
  const [activeLeaderboardTab, setActiveLeaderboardTab] = useState<'users' | 'schools' | 'teams'>('users');

  const stats = user?.stats;

  const loadData = async () => {
    try {
      const [qData, expData, achData, fragData] = await Promise.all([
        api.getDailyQuests().catch(() => []),
        api.getWeeklyExpedition().catch(() => null),
        api.getAchievements().catch(() => []),
        api.getFragments().catch(() => [])
      ]);
      setQuests(qData);
      setExpedition(expData);
      setAchievements(achData);
      if (Array.isArray(fragData) && fragData.length > 0) {
        setFragments(fragData.slice(0, 5));
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleClaim = async (questId: number) => {
    try {
      await api.claimQuest(questId);
      await Promise.all([loadData(), refreshUser()]);
    } catch (err: any) {
      alert(err.message || 'Ошибка получения награды');
    }
  };

  const currentFragment = fragments[currentIdx] || {
    id: 1,
    text_ba: 'Был донъяла иң ҡәҙерле нәмә — ваҡыт.',
    translations: [{ id: 1, translation_text: 'В этом мире самое ценное — время.' }],
    source_title: 'Башкирский эпос "Урал-батыр"',
    difficulty: 'medium',
  };

  const primaryTranslation = currentFragment.translations?.[0]?.translation_text || 'В этом мире самое ценное — время.';

  const handleConfirmTask = async () => {
    const tId = currentFragment.translations?.[0]?.id;
    if (tId) {
      try {
        await api.voteTranslation(tId, 'confirm');
        setConfirmedIds((prev) => [...prev, currentFragment.id]);
        setTaskNotice('✓ Перевод подтверждён! Получено +10 XP');
        await refreshUser();
        setTimeout(() => setTaskNotice(null), 3000);
      } catch (err: any) {
        setTaskNotice(err.message || 'Голос уже учтён');
        setTimeout(() => setTaskNotice(null), 3000);
      }
    }
  };

  const handlePlayAudio = () => {
    setIsPlayingAudio(true);
    if ('speechSynthesis' in window) {
      const utterance = new SpeechSynthesisUtterance(currentFragment.text_ba);
      utterance.lang = 'ru-RU';
      utterance.onend = () => setIsPlayingAudio(false);
      utterance.onerror = () => setIsPlayingAudio(false);
      window.speechSynthesis.speak(utterance);
    } else {
      setTimeout(() => setIsPlayingAudio(false), 1500);
    }
  };

  const xpProgress = stats
    ? Math.min(100, Math.round((stats.xp / Math.max(stats.next_level_xp, 1)) * 100))
    : 0;

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-12 select-none">
      {/* 1. SAPPHIRE HERO / USER WELCOME BANNER (Figma Style) */}
      <div className="bg-[#0C2340] rounded-3xl p-6 sm:p-8 text-white shadow-md relative overflow-hidden border border-slate-700/50">
        <div className="relative z-10 space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1E50A0]/60 text-xs font-semibold backdrop-blur mb-2 border border-blue-400/30">
                <Sparkles className="w-3.5 h-3.5 text-amber-400" />
                <span>Уровень {stats?.level || 1} — {stats?.level_title || 'Хранитель языка'}</span>
              </div>
              <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
                Хәйерле көн, {user?.full_name || user?.username}!
              </h1>
              <p className="text-slate-300 text-xs sm:text-sm mt-1">
                Школа: {user?.school_name || 'Индивидуальный Хранитель'} {user?.class_name ? `(${user.class_name})` : ''}
              </p>
            </div>

            <div className="flex items-center gap-3">
              {/* Daily Streak Pill */}
              <div className="flex items-center gap-1.5 px-3.5 py-2 rounded-2xl bg-white/10 border border-white/10">
                <Flame className="w-4 h-4 text-amber-400 fill-amber-400" />
                <span className="text-xs font-bold text-white">{stats?.current_streak || 1} дн. ударно</span>
              </div>

              {/* Kuray Coins Pill */}
              <div className="flex items-center gap-1.5 px-3.5 py-2 rounded-2xl bg-white/10 border border-white/10">
                <Coins className="w-4 h-4 text-amber-300" />
                <span className="text-xs font-bold text-amber-300">{stats?.currency || 0} Ҡурай</span>
              </div>
            </div>
          </div>

          {/* Progress to next level bar */}
          <div className="pt-2">
            <div className="flex justify-between text-xs text-slate-300 font-medium mb-1.5">
              <span>Прогресс до следующего уровня</span>
              <span className="font-bold text-white">
                {stats?.xp || 0} / {stats?.next_level_xp || 100} XP ({xpProgress}%)
              </span>
            </div>
            <div className="w-full h-2.5 bg-black/30 rounded-full overflow-hidden p-0.5">
              <div
                className="h-full bg-gradient-to-r from-[#1E50A0] to-[#3B82F6] rounded-full transition-all duration-500 shadow"
                style={{ width: `${xpProgress}%` }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* 2. DAILY TASK SECTION (Identical to Figma node 1:3 layout) */}
      <section className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="w-1 h-5 bg-[#1E50A0] rounded-full inline-block" />
            <h2 className="text-base sm:text-lg font-bold text-[#0C2340]">
              Ежедневное задание
            </h2>
          </div>

          <div className="flex items-center gap-1.5 text-xs font-semibold text-slate-500">
            <button
              onClick={() => setCurrentIdx((prev) => (prev - 1 + Math.max(fragments.length, 1)) % Math.max(fragments.length, 1))}
              className="p-1 rounded-lg border border-slate-200 hover:bg-slate-100 text-slate-600 transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="px-2">
              {currentIdx + 1} / {Math.max(fragments.length, 5)}
            </span>
            <button
              onClick={() => setCurrentIdx((prev) => (prev + 1) % Math.max(fragments.length, 1))}
              className="p-1 rounded-lg border border-slate-200 hover:bg-slate-100 text-slate-600 transition-colors"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>

        {taskNotice && (
          <div className="bg-[#EEF2FF] border border-[#1E50A0]/30 text-[#1E50A0] px-4 py-2 rounded-xl text-xs font-bold animate-fade-in">
            {taskNotice}
          </div>
        )}

        <div className="bg-white rounded-3xl border border-slate-200/90 p-5 sm:p-7 lg:p-8 shadow-sm">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center">
            {/* Left Content Column */}
            <div className="lg:col-span-7 space-y-4">
              <div className="space-y-1.5">
                <div className="text-xl sm:text-2xl font-black text-[#0C2340] leading-snug tracking-tight font-serif">
                  «{currentFragment.text_ba}»
                </div>
                <div className="text-sm sm:text-base text-slate-600 font-medium">
                  «{primaryTranslation}»
                </div>
              </div>

              {/* Source & Difficulty Badge */}
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#EEF2FF] text-[#1E50A0] text-xs font-bold">
                <span>{currentFragment.source_title || 'Башкирский эпос "Урал-батыр"'}</span>
                <span>•</span>
                <span>Сложность: {currentFragment.difficulty === 'hard' ? 'Высокая' : currentFragment.difficulty === 'easy' ? 'Базовая' : 'Средняя'}</span>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-wrap items-center gap-3 pt-2">
                <button
                  onClick={handleConfirmTask}
                  className={`px-5 py-2.5 rounded-full text-xs font-bold flex items-center gap-1.5 transition-all shadow-sm ${
                    confirmedIds.includes(currentFragment.id)
                      ? 'bg-emerald-600 text-white'
                      : 'bg-[#0C2340] hover:bg-[#1A4480] text-white transform hover:-translate-y-0.5'
                  }`}
                >
                  <Check className="w-3.5 h-3.5" />
                  <span>{confirmedIds.includes(currentFragment.id) ? 'Подтверждено ✓' : 'Подтвердить'}</span>
                </button>

                <Link
                  to="/task"
                  className="px-5 py-2.5 rounded-full bg-white hover:bg-blue-50 border border-[#1E50A0] text-[#1E50A0] text-xs font-bold flex items-center gap-1.5 transition-colors"
                >
                  <Edit3 className="w-3.5 h-3.5" />
                  <span>Исправить</span>
                </Link>

                <button
                  onClick={() => setCurrentIdx((prev) => (prev + 1) % Math.max(fragments.length, 1))}
                  className="px-4 py-2.5 rounded-full bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold transition-colors flex items-center gap-1"
                >
                  <span>Следующее</span>
                  <ArrowRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>

            {/* Right Instrument Audio Graphic Column */}
            <div className="lg:col-span-5 flex justify-center">
              <div className="relative w-full max-w-sm rounded-2xl overflow-hidden shadow-sm group border border-slate-100">
                <img
                  src={instrumentTask}
                  alt="Башкирский музыкальный инструмент"
                  className="w-full h-48 sm:h-52 object-cover transition-transform group-hover:scale-105 duration-300"
                />

                <button
                  onClick={handlePlayAudio}
                  disabled={isPlayingAudio}
                  className={`absolute inset-0 m-auto w-12 h-12 rounded-full bg-white/95 text-[#0C2340] shadow-xl flex items-center justify-center transition-all ${
                    isPlayingAudio ? 'scale-110 ring-4 ring-[#1E50A0]/40' : 'hover:scale-110 hover:bg-white'
                  }`}
                  title="Прослушать произношение"
                >
                  <Volume2 className={`w-5 h-5 ${isPlayingAudio ? 'text-[#1E50A0] animate-pulse' : 'text-[#0C2340]'}`} />
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 3. THREE-COLUMN BOTTOM GRID (Matching Figma node 1:3) */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {/* CARD 1: Сокровищница языка */}
        <div className="space-y-2 flex flex-col">
          <div className="flex items-center gap-2">
            <span className="w-1 h-4 bg-[#1E50A0] rounded-full inline-block" />
            <h3 className="text-sm font-bold text-[#0C2340]">
              Сокровищница языка
            </h3>
          </div>

          <div className="bg-white rounded-3xl border border-slate-200/90 p-5 shadow-sm flex-1 flex flex-col justify-between space-y-4">
            <div>
              <div className="text-2xl font-black text-[#0C2340] tracking-tight">
                ҡояш
              </div>
              <div className="text-xs text-slate-500 font-medium">
                солнце (корень башкирского языка)
              </div>

              <div className="flex flex-wrap items-center gap-1.5 mt-3">
                <span className="px-2 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] text-[10px] font-bold">
                  ba-ru
                </span>
                <span className="px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 text-[10px] font-bold flex items-center gap-1">
                  <Heart className="w-2.5 h-2.5 fill-emerald-600" />
                  <span>128 голосов</span>
                </span>
                <span className="px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 text-[10px] font-semibold">
                  Айгуль Х.
                </span>
              </div>

              <div className="mt-4 rounded-xl overflow-hidden border border-slate-100 shadow-inner">
                <img
                  src={treasurySunrise}
                  alt="Восход над горами Башкортостана"
                  className="w-full h-28 object-cover hover:scale-105 transition-transform duration-300"
                />
              </div>
            </div>

            <div className="pt-2 border-t border-slate-100">
              <Link
                to="/treasury"
                className="text-xs font-bold text-[#1E50A0] hover:text-[#0C2340] flex items-center justify-between group"
              >
                <span>Перейти в сокровищницу</span>
                <ArrowRight className="w-3.5 h-3.5 transform group-hover:translate-x-1 transition-transform" />
              </Link>
            </div>
          </div>
        </div>

        {/* CARD 2: Рейтинг сообщества */}
        <div className="space-y-2 flex flex-col">
          <div className="flex items-center gap-2">
            <span className="w-1 h-4 bg-[#1E50A0] rounded-full inline-block" />
            <h3 className="text-sm font-bold text-[#0C2340]">
              Рейтинг сообщества
            </h3>
          </div>

          <div className="bg-white rounded-3xl border border-slate-200/90 p-5 shadow-sm flex-1 flex flex-col justify-between space-y-4">
            <div>
              <div className="flex items-center gap-1 p-1 bg-slate-100 rounded-2xl mb-3">
                <button
                  onClick={() => setActiveLeaderboardTab('users')}
                  className={`flex-1 py-1 rounded-xl text-[11px] font-bold transition-all ${
                    activeLeaderboardTab === 'users'
                      ? 'bg-[#1E50A0] text-white shadow-sm'
                      : 'text-slate-500 hover:text-slate-900'
                  }`}
                >
                  Пользователи
                </button>
                <button
                  onClick={() => setActiveLeaderboardTab('schools')}
                  className={`flex-1 py-1 rounded-xl text-[11px] font-bold transition-all ${
                    activeLeaderboardTab === 'schools'
                      ? 'bg-[#1E50A0] text-white shadow-sm'
                      : 'text-slate-500 hover:text-slate-900'
                  }`}
                >
                  Школы
                </button>
                <button
                  onClick={() => setActiveLeaderboardTab('teams')}
                  className={`flex-1 py-1 rounded-xl text-[11px] font-bold transition-all ${
                    activeLeaderboardTab === 'teams'
                      ? 'bg-[#1E50A0] text-white shadow-sm'
                      : 'text-slate-500 hover:text-slate-900'
                  }`}
                >
                  Команды
                </button>
              </div>

              <div className="space-y-2.5">
                <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50 transition-colors">
                  <div className="flex items-center gap-2.5">
                    <span className="w-5 text-center font-black text-amber-500 text-sm">1</span>
                    <div className="w-7 h-7 rounded-full bg-blue-100 text-[#1E50A0] font-bold text-xs flex items-center justify-center">
                      АХ
                    </div>
                    <div>
                      <div className="font-bold text-slate-900 text-xs">Айгуль Х.</div>
                      <div className="text-[10px] text-slate-400">Уровень 12</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-[#1E50A0] text-xs">2 450 XP</div>
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50 transition-colors">
                  <div className="flex items-center gap-2.5">
                    <span className="w-5 text-center font-black text-slate-400 text-sm">2</span>
                    <div className="w-7 h-7 rounded-full bg-indigo-100 text-[#0C2340] font-bold text-xs flex items-center justify-center">
                      РС
                    </div>
                    <div>
                      <div className="font-bold text-slate-900 text-xs">Рамиль С.</div>
                      <div className="text-[10px] text-slate-400">Уровень 9</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-[#1E50A0] text-xs">1 980 XP</div>
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50 transition-colors">
                  <div className="flex items-center gap-2.5">
                    <span className="w-5 text-center font-black text-amber-700 text-sm">3</span>
                    <div className="w-7 h-7 rounded-full bg-slate-200 text-slate-800 font-bold text-xs flex items-center justify-center">
                      ЭГ
                    </div>
                    <div>
                      <div className="font-bold text-slate-900 text-xs">Элина Г.</div>
                      <div className="text-[10px] text-slate-400">Уровень 8</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-[#1E50A0] text-xs">1 720 XP</div>
                  </div>
                </div>
              </div>
            </div>

            <div className="pt-2 border-t border-slate-100">
              <Link
                to="/leaderboard"
                className="text-xs font-bold text-[#1E50A0] hover:text-[#0C2340] flex items-center justify-between group"
              >
                <span>Полный рейтинг</span>
                <ArrowRight className="w-3.5 h-3.5 transform group-hover:translate-x-1 transition-transform" />
              </Link>
            </div>
          </div>
        </div>

        {/* CARD 3: Мои достижения и ежедневные квесты */}
        <div className="space-y-2 flex flex-col">
          <div className="flex items-center gap-2">
            <span className="w-1 h-4 bg-[#1E50A0] rounded-full inline-block" />
            <h3 className="text-sm font-bold text-[#0C2340]">
              Мои достижения
            </h3>
          </div>

          <div className="bg-white rounded-3xl border border-slate-200/90 p-5 shadow-sm flex-1 flex flex-col justify-between space-y-4">
            <div className="space-y-3">
              {/* Circular Achievement Badges */}
              <div className="flex items-center justify-between px-1">
                <div
                  className="w-10 h-10 rounded-full bg-red-100 text-red-600 flex items-center justify-center shadow-sm cursor-pointer hover:scale-110 transition-transform"
                  title="Огонь знаний: 7 дней подряд"
                >
                  <Flame className="w-5 h-5 fill-red-500 text-red-500" />
                </div>
                <div
                  className="w-10 h-10 rounded-full bg-blue-100 text-[#1E50A0] flex items-center justify-center shadow-sm cursor-pointer hover:scale-110 transition-transform"
                  title="Полиглот: 50 фраз"
                >
                  <BookOpen className="w-5 h-5 text-[#1E50A0]" />
                </div>
                <div
                  className="w-10 h-10 rounded-full bg-indigo-100 text-[#0C2340] flex items-center justify-center shadow-sm cursor-pointer hover:scale-110 transition-transform"
                  title="Страж точности: 10 проверок"
                >
                  <Shield className="w-5 h-5 text-[#0C2340]" />
                </div>
                <div
                  className="w-10 h-10 rounded-full bg-amber-100 text-amber-600 flex items-center justify-center shadow-sm cursor-pointer hover:scale-110 transition-transform"
                  title="Мастер эпоса"
                >
                  <Trophy className="w-5 h-5 text-amber-600" />
                </div>
                <div
                  className="w-10 h-10 rounded-full bg-purple-100 text-purple-700 flex items-center justify-center shadow-sm cursor-pointer hover:scale-110 transition-transform"
                  title="Легенда Урала"
                >
                  <Crown className="w-5 h-5 text-purple-700" />
                </div>
              </div>

              {/* Active Milestone or Daily Quests Claim */}
              {quests.length > 0 && quests[0] ? (
                <div className="p-3.5 bg-slate-50 rounded-2xl border border-slate-100 space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="text-xs font-bold text-slate-900 truncate max-w-[150px]">
                      {quests[0].name}
                    </div>
                    {quests[0].completed && !quests[0].claimed ? (
                      <button
                        onClick={() => handleClaim(quests[0].id)}
                        className="text-[10px] font-bold text-white bg-[#1E50A0] hover:bg-[#2563EB] px-2.5 py-1 rounded-full shadow-sm animate-pulse"
                      >
                        Забрать +{quests[0].reward_xp} XP
                      </button>
                    ) : (
                      <span className="text-[10px] font-bold text-[#1E50A0] bg-[#EEF2FF] px-2 py-0.5 rounded-full">
                        {quests[0].progress} / {quests[0].target}
                      </span>
                    )}
                  </div>

                  <p className="text-[11px] text-slate-500 leading-tight">
                    {quests[0].description}
                  </p>

                  <div className="w-full h-2 bg-slate-200 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-[#1E50A0] rounded-full transition-all duration-500"
                      style={{
                        width: `${Math.min(100, Math.round((quests[0].progress / Math.max(quests[0].target, 1)) * 100))}%`
                      }}
                    />
                  </div>
                </div>
              ) : (
                <div className="p-3.5 bg-slate-50 rounded-2xl border border-slate-100 space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="text-xs font-bold text-slate-900">
                      Семидневный хранитель
                    </div>
                    <span className="text-[10px] font-bold text-[#1E50A0] bg-[#EEF2FF] px-2 py-0.5 rounded-full">
                      5 / 7 дней
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-500 leading-tight">
                    Подтверждайте переводы 7 дней подряд без перерыва
                  </p>
                  <div className="w-full h-2 bg-slate-200 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-[#1E50A0] rounded-full transition-all duration-500"
                      style={{ width: '71%' }}
                    />
                  </div>
                </div>
              )}
            </div>

            <div className="pt-2 border-t border-slate-100">
              <Link
                to="/achievements"
                className="text-xs font-bold text-[#1E50A0] hover:text-[#0C2340] flex items-center justify-between group"
              >
                <span>Все достижения</span>
                <ArrowRight className="w-3.5 h-3.5 transform group-hover:translate-x-1 transition-transform" />
              </Link>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
};
