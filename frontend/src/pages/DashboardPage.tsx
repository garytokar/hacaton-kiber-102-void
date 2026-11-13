import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { api } from '../api/client';
import { Fragment, PublicStats, LeaderboardUser, Achievement, DictionaryEntry, DailyQuest } from '../types';
import { OnboardingModal } from '../components/onboarding/OnboardingModal';
import { speakBashkir } from '../utils/speech';
import heroMountains from '../assets/figma/hero_mountains.png';
import instrumentTask from '../assets/figma/instrument_task.png';
import treasurySunrise from '../assets/figma/treasury_sunrise.png';
import {
  BookOpen,
  ArrowRight,
  Volume2,
  Headphones,
  Trophy,
  ExternalLink,
  Shield,
  Star,
  Mountain,
  Trees,
  Check
} from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { user, openAuthModal } = useAuth();

  // Real backend data states
  const [stats, setStats] = useState<PublicStats | null>(null);
  const [dailyFragment, setDailyFragment] = useState<Fragment | null>(null);
  const [leaders, setLeaders] = useState<LeaderboardUser[]>([]);
  const [archiveWord, setArchiveWord] = useState<DictionaryEntry | null>(null);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [dailyQuests, setDailyQuests] = useState<DailyQuest[]>([]);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);
  const [showOnboarding, setShowOnboarding] = useState(false);

  useEffect(() => {
    // 1. Public corpus stats
    api.getPublicStats()
      .then((data) => setStats(data))
      .catch(() => {});

    // 2. Daily task fragment preview
    api.getDailyFragment()
      .then((data) => setDailyFragment(data))
      .catch(() => {});

    // 3. Weekly leaders
    api.getLeaderboard('weekly', 'xp')
      .then((data) => {
        if (Array.isArray(data)) {
          setLeaders(data.slice(0, 3));
        }
      })
      .catch(() => {});

    // 4. Archive word from dictionary/treasury
    api.getDictionary()
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          setArchiveWord(data[0]);
        }
      })
      .catch(() => {});

    // 5. Achievements
    api.getAchievements()
      .then((data) => {
        if (Array.isArray(data)) {
          setAchievements(data);
        }
      })
      .catch(() => {});

    // 6. Daily Quests
    api.getDailyQuests()
      .then((data) => {
        if (Array.isArray(data)) {
          setDailyQuests(data);
        }
      })
      .catch(() => {});
  }, []);

  const handlePlayAudio = () => {
    if (!dailyFragment?.text_ba) return;
    setIsPlayingAudio(true);
    speakBashkir(dailyFragment.text_ba, {
      onStart: () => setIsPlayingAudio(true),
      onEnd: () => setIsPlayingAudio(false),
      onError: () => setIsPlayingAudio(false),
    });
  };

  const primaryTranslation = dailyFragment?.translations?.[0]?.translation_text || null;
  const userStreak = user?.stats?.current_streak ?? 0;
  const streakGoal = 7;
  const streakPercent = Math.min(100, Math.round((userStreak / streakGoal) * 100));

  return (
    <div className="space-y-6 max-w-7xl mx-auto pb-12 animate-fade-in">
      {/* 1. HERO BANNER: PANORAMIC SAPPHIRE WITH MOUNTAINS */}
      <section className="bg-gradient-to-r from-[#0C2340] via-[#0E2B52] to-[#143B6D] rounded-3xl p-6 sm:p-10 lg:p-12 text-white shadow-md relative overflow-hidden">
        {/* Right Mountain Art Overlay */}
        <div className="absolute right-0 top-0 bottom-0 w-full lg:w-3/5 pointer-events-none opacity-40 lg:opacity-75 mix-blend-screen overflow-hidden">
          <img
            src={heroMountains}
            alt="Уральские горы"
            className="w-full h-full object-cover object-right"
          />
        </div>

        <div className="relative z-10 grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
          {/* Left Text & CTA */}
          <div className="lg:col-span-8 space-y-5 max-w-xl">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/10 backdrop-blur-md border border-white/10 text-xs font-semibold text-slate-200">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
              <span>ЖИВОЙ КОРПУС БАШКИРСКОГО ЯЗЫКА</span>
            </div>

            <h1 className="text-3xl sm:text-4xl lg:text-5xl font-black tracking-tight leading-[1.12]">
              Сохраняй язык. <br />
              Создавай наследие.
            </h1>

            <p className="text-xs sm:text-sm text-slate-300 leading-relaxed max-w-lg">
              Язык остаётся живым, когда им делятся: переводите, сверяйте и сохраняйте башкирские тексты — от народных пословиц до аудиоархивов.
            </p>

            <div className="flex flex-wrap items-center gap-3 pt-2">
              <Link
                to="/task"
                className="px-6 py-3 rounded-full bg-white text-[#0C2340] hover:bg-slate-100 text-xs sm:text-sm font-black shadow-lg transition-all flex items-center gap-2 transform hover:-translate-y-0.5 cursor-pointer"
              >
                <span>Продолжить практику</span>
                <ArrowRight className="w-4 h-4" />
              </Link>

              {/* Requirement 7: Leads directly to /about (LandingPage) */}
              <Link
                to="/about"
                className="px-5 py-3 rounded-full bg-white/10 hover:bg-white/20 text-white border border-white/20 text-xs sm:text-sm font-semibold backdrop-blur-sm transition-colors cursor-pointer"
              >
                О платформе
              </Link>
            </div>
          </div>

          {/* Right Floating Stat Cards (Real data from API) */}
          <div className="lg:col-span-4 flex justify-end gap-3 self-start">
            <div className="p-4 rounded-2xl bg-[#0C2340]/70 backdrop-blur-md border border-white/10 text-left min-w-[130px] shadow-sm">
              <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                СЕГОДНЯ
              </div>
              <div className="text-xl sm:text-2xl font-black text-white mt-0.5">
                {stats ? stats.total_fragments.toLocaleString('ru-RU') : '—'}
              </div>
              <div className="text-xs text-slate-300 mt-0.5">
                слов в корпусе
              </div>
            </div>

            <div className="p-4 rounded-2xl bg-[#0C2340]/70 backdrop-blur-md border border-white/10 text-left min-w-[130px] shadow-sm">
              <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                СООБЩЕСТВО
              </div>
              <div className="text-xl sm:text-2xl font-black text-white mt-0.5">
                {stats ? stats.total_users.toLocaleString('ru-RU') : '—'}
              </div>
              <div className="text-xs text-slate-300 mt-0.5">
                хранителей
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 2. ФРАГМЕНТ ДНЯ (DAILY TASK PREVIEW - Single practice flow via /task) */}
      <section className="bg-white rounded-3xl border border-slate-200/90 p-6 sm:p-8 shadow-sm space-y-6">
        {/* Card Header */}
        <div className="flex items-center justify-between pb-4 border-b border-slate-100">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-[#0C2340] text-white flex items-center justify-center shadow-xs">
              <BookOpen className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                ФРАГМЕНТ ДНЯ • {dailyFragment?.category || 'ПОСЛОВИЦЫ'}
              </div>
              <h2 className="text-base sm:text-lg font-black text-[#0C2340]">
                Проверьте смысл, не только слова
              </h2>
            </div>
          </div>

          <Link
            to="/task"
            className="flex items-center gap-1.5 text-xs font-bold text-[#1E50A0] hover:text-[#0C2340] transition-colors"
          >
            <span>Вся практика</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>

        {/* Two-Column Task Preview Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
          {/* Left Column: Original & Translation */}
          <div className="lg:col-span-7 flex flex-col justify-between space-y-4">
            {/* Original Bashkir Card */}
            <div className="p-5 rounded-2xl bg-[#F8FAFC] border border-slate-200/80 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                  ОРИГИНАЛ • БАШҠОРТСА
                </span>
                <button
                  type="button"
                  onClick={handlePlayAudio}
                  disabled={isPlayingAudio}
                  className={`w-8 h-8 rounded-xl bg-white shadow-xs border border-slate-200 flex items-center justify-center transition-all cursor-pointer ${
                    isPlayingAudio ? 'ring-2 ring-[#1E50A0] text-[#1E50A0]' : 'text-slate-600 hover:text-[#0C2340]'
                  }`}
                  title="Прослушать башкирское произношение"
                >
                  <Volume2 className={`w-4 h-4 ${isPlayingAudio ? 'text-[#1E50A0]' : ''}`} />
                </button>
              </div>

              <div className="text-2xl sm:text-3xl font-black text-[#0C2340] leading-snug tracking-tight font-serif">
                {dailyFragment?.text_ba || 'Тыуған тел — халыҡтың рухи байлығы.'}
              </div>

              <div className="text-xs text-slate-500 font-medium flex items-center gap-1.5 pt-1">
                <span>📚</span>
                <span>{dailyFragment?.source_title || '«Башҡорт халыҡ мәҡәлдәре»'}</span>
              </div>
            </div>

            {/* Translation Card */}
            <div className="p-4 rounded-2xl bg-white border border-slate-200/80 space-y-1">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                ПЕРЕВОД • BA-RU
              </span>
              <div className="text-sm sm:text-base font-semibold text-slate-800">
                {primaryTranslation || 'Родной язык — духовное богатство народа.'}
              </div>
            </div>
          </div>

          {/* Right Column: Audio Box & Single Action to /task */}
          <div className="lg:col-span-5 flex flex-col justify-between space-y-3">
            {/* Audio Box */}
            <div
              onClick={handlePlayAudio}
              className="relative rounded-2xl overflow-hidden h-36 border border-slate-200/80 shadow-xs group cursor-pointer"
              title="Нажмите, чтобы прослушать башкирскую речь"
            >
              <img
                src={instrumentTask}
                alt="Башкирский инструмент думбыра"
                className="w-full h-full object-cover transition-transform group-hover:scale-105 duration-300"
              />

              <div className="absolute top-2.5 left-2.5 w-7 h-7 rounded-full bg-white/70 backdrop-blur-sm flex items-center justify-center text-[#0C2340] shadow-xs">
                <Headphones className="w-3.5 h-3.5" />
              </div>

              <div className="absolute top-2.5 right-2.5 w-7 h-7 rounded-full bg-white/70 backdrop-blur-sm flex items-center justify-center text-[#0C2340] shadow-xs">
                <Volume2 className={`w-3.5 h-3.5 ${isPlayingAudio ? 'text-[#1E50A0]' : ''}`} />
              </div>

              <div className="absolute bottom-0 inset-x-0 p-3 bg-gradient-to-t from-black/70 via-black/30 to-transparent text-white">
                <div className="text-xs font-bold leading-tight">Слушайте звучание</div>
                <div className="text-xs text-slate-200">Аудиофрагмент произношения</div>
              </div>
            </div>

            {/* Single Action Button leading to /task */}
            <div className="pt-2">
              <Link
                to="/task"
                className="w-full py-3.5 rounded-full bg-[#1E50A0] hover:bg-[#2563EB] text-white text-xs font-bold flex items-center justify-center gap-2 transition-all shadow-md cursor-pointer"
              >
                <span>Перейти к выполнению задания</span>
                <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* 3. THREE-COLUMN BOTTOM GRID */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {/* CARD 1: Сокровищница языка */}
        <div className="bg-white rounded-3xl border border-slate-200/90 p-6 shadow-sm flex flex-col justify-between space-y-4">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <div>
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                  СЛОВО ИЗ АРХИВА
                </span>
                <h3 className="text-sm font-bold text-[#0C2340]">
                  Сокровищница языка
                </h3>
              </div>
              <Link
                to="/treasury"
                className="w-8 h-8 rounded-full bg-slate-50 hover:bg-slate-100 flex items-center justify-center text-[#1E50A0] transition-colors"
                title="Открыть сокровищницу"
              >
                <ExternalLink className="w-4 h-4" />
              </Link>
            </div>

            <div className="pt-4 space-y-2">
              <div className="text-3xl font-black text-[#0C2340] tracking-tight">
                {archiveWord ? archiveWord.word : 'ҡояш'}
              </div>
              <div className="text-xs text-slate-500 font-medium">
                {archiveWord ? archiveWord.translation : 'солнце'}
              </div>

              {archiveWord?.context && (
                <div className="text-xs text-slate-400 italic pt-0.5">
                  «{archiveWord.context}»
                </div>
              )}

              <div className="mt-4 rounded-2xl overflow-hidden border border-slate-100 shadow-inner">
                <img
                  src={treasurySunrise}
                  alt="Восход над горами"
                  className="w-full h-28 object-cover hover:scale-105 transition-transform duration-300"
                />
              </div>
            </div>
          </div>
        </div>

        {/* CARD 2: Лидеры недели (Real Leaderboard) */}
        <div className="bg-white rounded-3xl border border-slate-200/90 p-6 shadow-sm flex flex-col justify-between space-y-4">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <div>
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                  СИЛА СООБЩЕСТВА
                </span>
                <h3 className="text-sm font-bold text-[#0C2340]">
                  Лидеры недели
                </h3>
              </div>
              <Link
                to="/leaderboard"
                className="w-8 h-8 rounded-full bg-amber-50 text-amber-600 flex items-center justify-center"
                title="Все лидеры"
              >
                <Trophy className="w-4 h-4" />
              </Link>
            </div>

            <div className="pt-3 space-y-2.5">
              {leaders.length > 0 ? (
                leaders.map((leader, idx) => (
                  <div key={leader.user_id || idx} className="flex items-center justify-between text-xs p-2.5 rounded-xl bg-slate-50/70">
                    <div className="flex items-center gap-2.5">
                      <span className={`w-4 font-bold text-xs ${idx === 0 ? 'text-amber-500' : idx === 1 ? 'text-slate-400' : 'text-amber-700'}`}>
                        {idx + 1}
                      </span>
                      <div className="w-7 h-7 rounded-full bg-blue-100 text-[#1E50A0] font-bold text-xs flex items-center justify-center">
                        {(leader.full_name || leader.username || 'Х').slice(0, 2).toUpperCase()}
                      </div>
                      <span className="font-semibold text-slate-900 text-xs truncate max-w-[120px]">
                        {leader.full_name || leader.username}
                      </span>
                    </div>
                    <span className="font-bold text-[#1E50A0] text-xs">
                      {leader.xp.toLocaleString('ru-RU')} XP
                    </span>
                  </div>
                ))
              ) : (
                <div className="text-xs text-slate-400 py-6 text-center">
                  Пока нет данных лидеров
                </div>
              )}
            </div>
          </div>
        </div>

        {/* CARD 3: Достижения Хранителя (Real User Progress) */}
        <div className="bg-[#0C2340] rounded-3xl p-6 text-white shadow-sm flex flex-col justify-between space-y-4">
          <div>
            <div className="pb-3 border-b border-white/10">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                ДОСТИЖЕНИЯ ХРАНИТЕЛЯ
              </span>
              <h3 className="text-sm font-bold text-white mt-0.5">
                Коллекция
              </h3>
            </div>

            <div className="pt-4 space-y-4">
              {/* Row of 5 blue circular badges */}
              <div className="flex items-center justify-between">
                <div className="w-9 h-9 rounded-full bg-[#1E50A0]/60 text-white flex items-center justify-center shadow-xs">
                  <BookOpen className="w-4 h-4 text-white" />
                </div>
                <div className="w-9 h-9 rounded-full bg-[#1E50A0]/60 text-white flex items-center justify-center shadow-xs">
                  <Star className="w-4 h-4 text-white" />
                </div>
                <div className="w-9 h-9 rounded-full bg-[#1E50A0]/60 text-white flex items-center justify-center shadow-xs">
                  <Mountain className="w-4 h-4 text-white" />
                </div>
                <div className="w-9 h-9 rounded-full bg-[#1E50A0]/60 text-white flex items-center justify-center shadow-xs">
                  <Trees className="w-4 h-4 text-white" />
                </div>
                <div className="w-9 h-9 rounded-full bg-[#1E50A0]/60 text-white flex items-center justify-center shadow-xs">
                  <Shield className="w-4 h-4 text-white" />
                </div>
              </div>

              {/* Progress Card */}
              {user ? (
                <div className="p-3.5 rounded-2xl bg-white/10 border border-white/10 space-y-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-white">Семидневный хранитель</span>
                    <span className="text-xs font-semibold text-amber-300 bg-amber-400/20 px-2 py-0.5 rounded-full">
                      {userStreak} / {streakGoal} дней
                    </span>
                  </div>
                  <p className="text-xs text-slate-300 leading-tight">
                    Выполняйте задания 7 дней подряд
                  </p>
                  <div className="w-full h-1.5 bg-white/20 rounded-full overflow-hidden">
                    <div className="h-full bg-blue-400 rounded-full" style={{ width: `${streakPercent}%` }} />
                  </div>
                </div>
              ) : (
                <div className="p-3.5 rounded-2xl bg-white/10 border border-white/10 text-center space-y-2">
                  <p className="text-xs text-slate-300">
                    Войдите в аккаунт, чтобы отслеживать прогресс и получать награды
                  </p>
                  <button
                    type="button"
                    onClick={() => openAuthModal('/achievements')}
                    className="inline-block px-4 py-1.5 rounded-full bg-white text-[#0C2340] text-xs font-bold hover:bg-slate-100 transition-colors cursor-pointer"
                  >
                    Войти
                  </button>
                </div>
              )}

              {/* Daily Quest Highlight */}
              {dailyQuests.length > 0 && (
                <div className="p-3.5 rounded-2xl bg-white/10 border border-white/10 space-y-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-white flex items-center gap-1.5 truncate max-w-[170px]">
                      <span className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0" />
                      <span className="truncate">{dailyQuests[0].name}</span>
                    </span>
                    <span className="text-[11px] font-bold text-amber-300 bg-amber-400/20 px-2 py-0.5 rounded-full shrink-0">
                      +{dailyQuests[0].reward_xp} XP
                    </span>
                  </div>
                  <p className="text-xs text-slate-300 leading-tight">
                    {dailyQuests[0].description}
                  </p>
                  <div className="flex items-center justify-between text-[11px] text-slate-300 pt-0.5">
                    <span>
                      {Math.min(dailyQuests[0].progress, dailyQuests[0].target)} / {dailyQuests[0].target}
                    </span>
                    {dailyQuests[0].completed && (
                      <span className="text-emerald-400 font-bold flex items-center gap-1">
                        <Check className="w-3 h-3" /> Выполнен
                      </span>
                    )}
                  </div>
                  <div className="w-full h-1.5 bg-white/20 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-300 ${
                        dailyQuests[0].completed ? 'bg-emerald-400' : 'bg-blue-400'
                      }`}
                      style={{
                        width: `${Math.min(100, Math.round((dailyQuests[0].progress / dailyQuests[0].target) * 100))}%`,
                      }}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* Onboarding Modal retained for explicit usage */}
      <OnboardingModal
        isOpen={showOnboarding}
        onClose={() => setShowOnboarding(false)}
      />
    </div>
  );
};
