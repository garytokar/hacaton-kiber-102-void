import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Fragment, LeaderboardUser, SchoolLeaderboardEntry, TeamLeaderboardEntry } from '../types';
import heroMountains from '../assets/figma/hero_mountains.png';
import instrumentTask from '../assets/figma/instrument_task.png';
import treasurySunrise from '../assets/figma/treasury_sunrise.png';
import {
  Volume2,
  Check,
  Edit3,
  ChevronLeft,
  ChevronRight,
  ArrowRight,
  Flame,
  BookOpen,
  Shield,
  Trophy,
  Crown,
  Heart,
  Sparkles,
  Share2,
  Info
} from 'lucide-react';

export const LandingPage: React.FC = () => {
  const { user, quickLogin } = useAuth();
  const navigate = useNavigate();

  // Daily Tasks state
  const [fragments, setFragments] = useState<Fragment[]>([]);
  const [currentIdx, setCurrentIdx] = useState(0);
  const [confirmedIds, setConfirmedIds] = useState<number[]>([]);
  const [taskMessage, setTaskMessage] = useState<string | null>(null);

  // Community Leaderboard state
  const [activeTab, setActiveTab] = useState<'users' | 'schools' | 'teams'>('users');
  const [topUsers, setTopUsers] = useState<LeaderboardUser[]>([]);
  const [topSchools, setTopSchools] = useState<SchoolLeaderboardEntry[]>([]);
  const [topTeams, setTopTeams] = useState<TeamLeaderboardEntry[]>([]);

  // Treasury featured item state
  const [featuredWord, setFeaturedWord] = useState({
    word: 'ҡояш',
    translation: 'солнце',
    votes: 128,
    author: 'Айгуль Х.',
    category: 'Природа и космос'
  });

  // Sound playing state
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  useEffect(() => {
    // 1. Fetch public fragments for the daily task slider
    api.getFragments()
      .then((res) => {
        if (Array.isArray(res) && res.length > 0) {
          setFragments(res.slice(0, 5));
        }
      })
      .catch((err) => console.error('Failed to load fragments:', err));

    // 2. Fetch Leaderboard data
    api.getLeaderboard('all_time', 'xp')
      .then((users) => setTopUsers(users.slice(0, 3)))
      .catch(() => {});

    api.getSchoolLeaderboard()
      .then((schools) => setTopSchools(schools.slice(0, 3)))
      .catch(() => {});

    api.getTeamLeaderboard()
      .then((teams) => setTopTeams(teams.slice(0, 3)))
      .catch(() => {});

    // 3. Fetch Treasury item for card 1
    api.getTreasury()
      .then((res) => {
        if (res.items && res.items.length > 0) {
          const item = res.items[0];
          setFeaturedWord({
            word: item.text_ba,
            translation: item.translation_text,
            votes: item.votes_count || 128,
            author: item.user_name || 'Айгуль Х.',
            category: item.category || 'Наследие'
          });
        }
      })
      .catch(() => {});
  }, []);

  const currentFragment = fragments[currentIdx] || {
    id: 1,
    text_ba: 'Был донъяла иң ҡәҙерле нәмә — ваҡыт.',
    translations: [{ id: 1, translation_text: 'В этом мире самое ценное — время.' }],
    source_title: 'Башкирский эпос "Урал-батыр"',
    difficulty: 'medium',
  };

  const primaryTranslation = currentFragment.translations?.[0]?.translation_text || 'В этом мире самое ценное — время.';

  const handleNextFragment = () => {
    if (fragments.length > 0) {
      setCurrentIdx((prev) => (prev + 1) % fragments.length);
    }
  };

  const handlePrevFragment = () => {
    if (fragments.length > 0) {
      setCurrentIdx((prev) => (prev - 1 + fragments.length) % fragments.length);
    }
  };

  const handleConfirmTask = async () => {
    if (!user) {
      setTaskMessage('Войдите, чтобы сохранить прогресс и получить +10 XP!');
      setTimeout(() => setTaskMessage(null), 3500);
      return;
    }
    const tId = currentFragment.translations?.[0]?.id;
    if (tId) {
      try {
        await api.voteTranslation(tId, 'confirm');
        setConfirmedIds((prev) => [...prev, currentFragment.id]);
        setTaskMessage('✓ Перевод подтверждён! Получено +10 XP');
        setTimeout(() => setTaskMessage(null), 3000);
      } catch (err: any) {
        setTaskMessage(err.message || 'Голос уже учтён');
        setTimeout(() => setTaskMessage(null), 3000);
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

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-12 select-none">
      {/* 1. HERO SECTION (Matching Figma node 1:3) */}
      <section className="bg-white rounded-3xl border border-slate-200/90 p-6 sm:p-8 lg:p-10 shadow-sm relative overflow-hidden">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
          {/* Left Text & CTA */}
          <div className="lg:col-span-7 space-y-5">
            <h1 className="text-3xl sm:text-4xl lg:text-5xl font-black text-[#0C2340] tracking-tight leading-[1.15]">
              Каждое слово <br />
              сохраняет язык
            </h1>

            <p className="text-slate-600 text-sm sm:text-base leading-relaxed max-w-xl">
              Интерактивная платформа для сохранения и развития башкирского языка через соревновательное обучение и машинный перевод.
            </p>

            <div className="flex flex-wrap items-center gap-3 pt-2">
              <Link
                to="/task"
                className="px-6 py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs sm:text-sm font-bold shadow-md hover:shadow-lg transition-all flex items-center gap-2 transform hover:-translate-y-0.5"
              >
                <span>Начать задание</span>
                <ArrowRight className="w-4 h-4" />
              </Link>

              <a
                href="#about"
                onClick={(e) => {
                  e.preventDefault();
                  alert('«ЙӘНЛӘНДЕР» — цифровая инициатива хакатона по оцифровке, проверке и сохранению языкового наследия Башкортостана.');
                }}
                className="px-5 py-3 rounded-full bg-white hover:bg-slate-50 border border-slate-300 text-[#0C2340] text-xs sm:text-sm font-bold transition-colors"
              >
                О проекте
              </a>
            </div>
          </div>

          {/* Right Ural Mountains Art */}
          <div className="lg:col-span-5 flex justify-center lg:justify-end">
            <div className="relative max-w-sm sm:max-w-md w-full">
              <img
                src={heroMountains}
                alt="Уральские горы Башкортостана"
                className="w-full h-auto object-cover rounded-2xl shadow-sm"
              />
            </div>
          </div>
        </div>
      </section>

      {/* 2. DAILY TASK SECTION (Matching Figma node 1:3) */}
      <section className="space-y-3">
        {/* Section Header with Blue Vertical Line and Pagination */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="w-1 h-5 bg-[#1E50A0] rounded-full inline-block" />
            <h2 className="text-base sm:text-lg font-bold text-[#0C2340]">
              Ежедневное задание
            </h2>
          </div>

          {/* Pagination Controls */}
          <div className="flex items-center gap-1.5 text-xs font-semibold text-slate-500">
            <button
              onClick={handlePrevFragment}
              className="p-1 rounded-lg border border-slate-200 hover:bg-slate-100 text-slate-600 transition-colors"
              title="Предыдущее задание"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="px-2">
              {currentIdx + 1} / {Math.max(fragments.length, 5)}
            </span>
            <button
              onClick={handleNextFragment}
              className="p-1 rounded-lg border border-slate-200 hover:bg-slate-100 text-slate-600 transition-colors"
              title="Следующее задание"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Task Notification Toast */}
        {taskMessage && (
          <div className="bg-[#EEF2FF] border border-[#1E50A0]/30 text-[#1E50A0] px-4 py-2 rounded-xl text-xs font-bold animate-fade-in flex items-center justify-between">
            <span>{taskMessage}</span>
            {!user && (
              <button
                onClick={() => quickLogin('student1')}
                className="underline hover:text-[#0C2340] ml-2"
              >
                Быстрый демо-вход
              </button>
            )}
          </div>
        )}

        {/* Main Task Card */}
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
                  onClick={handleNextFragment}
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

                {/* Circular Sound Play Button in Center */}
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
              {/* Word & translation */}
              <div className="text-2xl font-black text-[#0C2340] tracking-tight">
                {featuredWord.word}
              </div>
              <div className="text-xs text-slate-500 font-medium">
                {featuredWord.translation}
              </div>

              {/* Tags */}
              <div className="flex flex-wrap items-center gap-1.5 mt-3">
                <span className="px-2 py-0.5 rounded-full bg-[#EEF2FF] text-[#1E50A0] text-[10px] font-bold">
                  ba-ru
                </span>
                <span className="px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 text-[10px] font-bold flex items-center gap-1">
                  <Heart className="w-2.5 h-2.5 fill-emerald-600" />
                  <span>{featuredWord.votes} голосов</span>
                </span>
                <span className="px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 text-[10px] font-semibold">
                  {featuredWord.author}
                </span>
              </div>

              {/* Sunrise graphic */}
              <div className="mt-4 rounded-xl overflow-hidden border border-slate-100 shadow-inner">
                <img
                  src={treasurySunrise}
                  alt="Восход над горами Башкортостана"
                  className="w-full h-28 object-cover hover:scale-105 transition-transform duration-300"
                />
              </div>
            </div>

            {/* Bottom link */}
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
              {/* Tabs */}
              <div className="flex items-center gap-1 p-1 bg-slate-100 rounded-2xl mb-3">
                <button
                  onClick={() => setActiveTab('users')}
                  className={`flex-1 py-1 rounded-xl text-[11px] font-bold transition-all ${
                    activeTab === 'users'
                      ? 'bg-[#1E50A0] text-white shadow-sm'
                      : 'text-slate-500 hover:text-slate-900'
                  }`}
                >
                  Пользователи
                </button>
                <button
                  onClick={() => setActiveTab('schools')}
                  className={`flex-1 py-1 rounded-xl text-[11px] font-bold transition-all ${
                    activeTab === 'schools'
                      ? 'bg-[#1E50A0] text-white shadow-sm'
                      : 'text-slate-500 hover:text-slate-900'
                  }`}
                >
                  Школы
                </button>
                <button
                  onClick={() => setActiveTab('teams')}
                  className={`flex-1 py-1 rounded-xl text-[11px] font-bold transition-all ${
                    activeTab === 'teams'
                      ? 'bg-[#1E50A0] text-white shadow-sm'
                      : 'text-slate-500 hover:text-slate-900'
                  }`}
                >
                  Команды
                </button>
              </div>

              {/* Items List */}
              <div className="space-y-2.5">
                {activeTab === 'users' && (
                  <>
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
                  </>
                )}

                {activeTab === 'schools' && (
                  <>
                    <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-amber-500">1.</span>
                        <span className="font-bold text-slate-800 truncate max-w-[130px]">Башкирская гимназия №1</span>
                      </div>
                      <span className="font-bold text-[#1E50A0]">8 920 XP</span>
                    </div>
                    <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-slate-400">2.</span>
                        <span className="font-bold text-slate-800 truncate max-w-[130px]">Лицей №153 г. Уфы</span>
                      </div>
                      <span className="font-bold text-[#1E50A0]">6 410 XP</span>
                    </div>
                    <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-amber-700">3.</span>
                        <span className="font-bold text-slate-800 truncate max-w-[130px]">Гимназия №39 им. Файзи</span>
                      </div>
                      <span className="font-bold text-[#1E50A0]">5 180 XP</span>
                    </div>
                  </>
                )}

                {activeTab === 'teams' && (
                  <>
                    <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-amber-500">1.</span>
                        <span className="font-bold text-slate-800 truncate max-w-[130px]">Урал Батырҙары</span>
                      </div>
                      <span className="font-bold text-[#1E50A0]">5 420 XP</span>
                    </div>
                    <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-slate-400">2.</span>
                        <span className="font-bold text-slate-800 truncate max-w-[130px]">Ағиҙел Ҡарлуғастары</span>
                      </div>
                      <span className="font-bold text-[#1E50A0]">4 190 XP</span>
                    </div>
                    <div className="flex items-center justify-between text-xs p-1.5 rounded-xl hover:bg-slate-50">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-amber-700">3.</span>
                        <span className="font-bold text-slate-800 truncate max-w-[130px]">Торатау Бөркөттәре</span>
                      </div>
                      <span className="font-bold text-[#1E50A0]">3 860 XP</span>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Bottom link */}
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

        {/* CARD 3: Мои достижения */}
        <div className="space-y-2 flex flex-col">
          <div className="flex items-center gap-2">
            <span className="w-1 h-4 bg-[#1E50A0] rounded-full inline-block" />
            <h3 className="text-sm font-bold text-[#0C2340]">
              Мои достижения
            </h3>
          </div>

          <div className="bg-white rounded-3xl border border-slate-200/90 p-5 shadow-sm flex-1 flex flex-col justify-between space-y-4">
            <div className="space-y-4">
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

              {/* Active Milestone Card */}
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

                {/* Progress bar */}
                <div className="w-full h-2 bg-slate-200 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-[#1E50A0] rounded-full transition-all duration-500"
                    style={{ width: '71%' }}
                  />
                </div>
              </div>
            </div>

            {/* Bottom link */}
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
