import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { useLang } from '../context/LanguageContext';
import { Fragment } from '../types';
import heroMountains from '../assets/figma/hero_mountains.png';
import instrumentTask from '../assets/figma/instrument_task.png';
import { BashkirOrnament } from '../components/ui/BashkirOrnament';
import { AudioSpeakerButton } from '../components/ui/AudioSpeakerButton';
import { speakBashkir } from '../utils/speech';
import {
  Sparkles,
  ArrowRight,
  Zap,
  CheckCircle2,
  XCircle,
  Volume2,
  ShoppingBag,
  Users2,
  BookOpen,
  BarChart3,
  Globe,
  GraduationCap,
  BookmarkCheck,
  ChevronRight,
  LogIn,
  Check,
  Award,
  Flame,
  Shield,
  Layers,
  HelpCircle
} from 'lucide-react';

export const LandingPage: React.FC = () => {
  const { user, quickLogin } = useAuth();
  const { lang, toggleLang, t } = useLang();
  const navigate = useNavigate();

  const [fragments, setFragments] = useState<Fragment[]>([]);
  const [currentIdx, setCurrentIdx] = useState(0);
  const [confirmed, setConfirmed] = useState(false);
  const [isStarting, setIsStarting] = useState(false);

  useEffect(() => {
    api.getFragments()
      .then((res) => {
        if (Array.isArray(res) && res.length > 0) {
          setFragments(res.slice(0, 5));
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

  // One-click primary action: leads directly to the platform dashboard
  const handleProceed = async (demoEmail: string = 'student_timur@yanlander.ru') => {
    setIsStarting(true);
    try {
      if (user) {
        navigate('/dashboard');
      } else {
        await quickLogin(demoEmail);
        navigate('/dashboard');
      }
    } catch {
      navigate('/dashboard');
    } finally {
      setIsStarting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#F4F6F9] text-[#0C2340] font-sans antialiased selection:bg-[#1E50A0] selection:text-white">
      {/* 1. STANDALONE CLEAN HEADER (NO DASHBOARD SIDEBAR/NAVBAR) */}
      <header className="sticky top-0 z-50 bg-white/90 backdrop-blur-md border-b border-slate-200/80 px-4 sm:px-8 py-3.5 transition-all">
        <div className="max-w-7xl mx-auto flex items-center justify-between gap-4">
          {/* Logo & Brand */}
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-[#0C2340] flex items-center justify-center text-white shadow-sm ring-2 ring-[#0C2340]/10">
              <BashkirOrnament className="w-6 h-6 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-1.5">
                <span className="text-lg font-black tracking-wider text-[#0C2340]">ЙӘНЛӘНДЕР</span>
              </div>
              <p className="text-[11px] font-semibold text-slate-500 hidden sm:block">
                Платформа сохранения башкирского языка
              </p>
            </div>
          </div>

          {/* Quick Nav Anchors */}
          <nav className="hidden md:flex items-center gap-6 text-xs font-bold text-slate-600">
            <a href="#about" className="hover:text-[#1E50A0] transition-colors">О проекте</a>
            <a href="#how-it-works" className="hover:text-[#1E50A0] transition-colors">Как устроен</a>
            <a href="#comparison" className="hover:text-[#1E50A0] transition-colors">Плюсы и минусы</a>
            <a href="#features" className="hover:text-[#1E50A0] transition-colors">Возможности</a>
          </nav>

          {/* Right Action Controls */}
          <div className="flex items-center gap-3">
            {/* Language toggle pill */}
            <button
              type="button"
              onClick={toggleLang}
              className="px-3 py-1.5 rounded-full bg-slate-100 hover:bg-slate-200 text-xs font-bold text-slate-700 transition-colors flex items-center gap-1.5"
              title="Переключить язык (RU / BA)"
            >
              <Globe className="w-3.5 h-3.5 text-[#1E50A0]" />
              <span>{lang === 'ru' ? 'РУС' : 'БАШ'}</span>
            </button>

            {user ? (
              <button
                type="button"
                onClick={() => navigate('/dashboard')}
                className="px-5 py-2 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs sm:text-sm font-bold shadow-md hover:shadow-lg transition-all flex items-center gap-2"
              >
                <span>В дашборд</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            ) : (
              <div className="flex items-center gap-2">
                <Link
                  to="/login"
                  className="px-4 py-2 rounded-full text-slate-700 hover:text-[#0C2340] text-xs font-bold transition-colors hidden sm:inline-block"
                >
                  Войти
                </Link>
                <button
                  type="button"
                  onClick={() => handleProceed('student_timur@yanlander.ru')}
                  disabled={isStarting}
                  className="px-5 py-2.5 rounded-full bg-gradient-to-r from-[#0C2340] to-[#1E50A0] hover:from-[#13335e] hover:to-[#2868cc] text-white text-xs sm:text-sm font-black shadow-md hover:shadow-lg transition-all flex items-center gap-2 cursor-pointer"
                >
                  <Zap className="w-3.5 h-3.5 text-amber-400 fill-amber-400" />
                  <span>{isStarting ? 'Вход...' : 'Приступить'}</span>
                  <ArrowRight className="w-3.5 h-3.5" />
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* 2. MAIN HERO SECTION */}
      <main className="max-w-7xl mx-auto px-4 sm:px-8 py-8 sm:py-12 space-y-16">
        <section id="about" className="bg-white rounded-3xl border border-slate-200/90 p-6 sm:p-10 lg:p-14 shadow-sm relative overflow-hidden">
          <div className="absolute -top-32 -right-32 w-96 h-96 bg-blue-100/50 rounded-full blur-3xl pointer-events-none" />
          <div className="absolute -bottom-32 -left-32 w-96 h-96 bg-indigo-100/40 rounded-full blur-3xl pointer-events-none" />

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-center relative z-10">
            {/* Left Copy & Actions */}
            <div className="lg:col-span-7 space-y-6">
              <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-[#0C2340] text-amber-300 text-xs font-black tracking-wide shadow-sm">
                <Sparkles className="w-3.5 h-3.5 text-amber-400" />
                <span>ПЛАТФОРМА КРАУДСОРСИНГА • СОХРАНЕНИЕ ЯЗЫКА</span>
              </div>

              <h1 className="text-3xl sm:text-5xl lg:text-6xl font-black text-[#0C2340] tracking-tight leading-[1.1]">
                Сохраняй язык. <br />
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#1E50A0] via-[#0C2340] to-[#1E50A0]">
                  Создавай наследие.
                </span>
              </h1>

              <p className="text-slate-600 text-sm sm:text-base leading-relaxed max-w-xl">
                <strong className="text-slate-900">«ЙӘНЛӘНДЕР»</strong> — первая интерактивная платформа Республики Башкортостан для сохранения, изучения и краудсорсинг-верификации башкирского языка.
                Мы объединяем аутентичную озвучку фонем с кураем, школьные лиги классов, геймификацию и Лавку Хранителя.
              </p>

              {/* PRIMARY PROCEED BUTTON */}
              <div className="space-y-4 pt-2">
                <div className="flex flex-wrap items-center gap-3">
                  <button
                    type="button"
                    onClick={() => handleProceed('student_timur@yanlander.ru')}
                    disabled={isStarting}
                    className="px-8 py-4 rounded-full bg-gradient-to-r from-[#0C2340] via-[#153a6b] to-[#1E50A0] hover:from-[#13335e] hover:to-[#2868cc] text-white text-sm sm:text-base font-black shadow-xl hover:shadow-2xl transition-all flex items-center gap-3 transform hover:-translate-y-0.5 cursor-pointer disabled:opacity-70 group"
                  >
                    <Zap className="w-5 h-5 text-amber-400 fill-amber-400 group-hover:scale-110 transition-transform" />
                    <span>{isStarting ? 'Загрузка дашборда...' : 'Приступить к изучению (Дашборд)'}</span>
                    <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                  </button>

                  <a
                    href="#comparison"
                    className="px-5 py-4 rounded-full bg-slate-100 hover:bg-slate-200 border border-slate-200 text-[#0C2340] text-xs sm:text-sm font-bold transition-colors"
                  >
                    Смотреть плюсы и минусы
                  </a>
                </div>

                {/* Quick Judge Login Chips */}
                <div className="flex flex-wrap items-center gap-2 pt-1 text-xs">
                  <span className="text-slate-500 font-semibold text-[11px] flex items-center gap-1">
                    <GraduationCap className="w-3.5 h-3.5 text-[#1E50A0]" />
                    Быстрый вход для жюри:
                  </span>
                  <button
                    type="button"
                    onClick={() => handleProceed('student_timur@yanlander.ru')}
                    className="px-3 py-1 rounded-lg bg-blue-50 hover:bg-blue-100 text-[#1E50A0] font-bold border border-blue-200 transition-colors cursor-pointer"
                  >
                    Ученик
                  </button>
                  <button
                    type="button"
                    onClick={() => handleProceed('teacher_aygul@yanlander.ru')}
                    className="px-3 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold border border-slate-200 transition-colors cursor-pointer"
                  >
                    Учитель
                  </button>
                  <button
                    type="button"
                    onClick={() => handleProceed('expert@yanlander.ru')}
                    className="px-3 py-1 rounded-lg bg-purple-50 hover:bg-purple-100 text-purple-700 font-bold border border-purple-200 transition-colors cursor-pointer"
                  >
                    Эксперт
                  </button>
                </div>
              </div>
            </div>

            {/* Right Ural Mountains Artwork */}
            <div className="lg:col-span-5 flex justify-center lg:justify-end">
              <div className="relative max-w-sm sm:max-w-md w-full">
                <div className="absolute -inset-1.5 bg-gradient-to-r from-amber-400 via-blue-500 to-indigo-600 rounded-3xl blur-md opacity-25" />
                <div className="relative rounded-2xl overflow-hidden border border-slate-200/90 shadow-xl bg-white">
                  <img
                    src={heroMountains}
                    alt="Уральские горы Башкортостана"
                    className="w-full h-auto object-cover"
                  />
                  <div className="p-4 bg-white/95 backdrop-blur-sm border-t border-slate-100 flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                      <div className="w-8 h-8 rounded-xl bg-[#0C2340] text-amber-400 flex items-center justify-center font-bold text-xs shadow-sm">
                        <BashkirOrnament className="w-5 h-5 text-white" />
                      </div>
                      <div>
                        <div className="text-xs font-black text-[#0C2340]">Эпос «Урал-батыр»</div>
                        <div className="text-[10px] text-slate-500">Культурное наследие Башкортостана</div>
                      </div>
                    </div>
                    <span className="text-[10px] font-extrabold px-2.5 py-1 rounded-full bg-emerald-100 text-emerald-800">
                      Живой язык
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Key Metrics Strip */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-8 mt-10 border-t border-slate-100">
            <div className="p-3.5 bg-slate-50/80 rounded-2xl border border-slate-100 text-center">
              <div className="text-2xl font-black text-[#0C2340]">34+</div>
              <div className="text-[11px] text-slate-500 font-bold">Фрагментов в корпусе</div>
            </div>
            <div className="p-3.5 bg-slate-50/80 rounded-2xl border border-slate-100 text-center">
              <div className="text-2xl font-black text-[#1E50A0]">31</div>
              <div className="text-[11px] text-slate-500 font-bold">Активных Хранителей</div>
            </div>
            <div className="p-3.5 bg-slate-50/80 rounded-2xl border border-slate-100 text-center">
              <div className="text-2xl font-black text-emerald-600">5</div>
              <div className="text-[11px] text-slate-500 font-bold">Школ в экспедициях</div>
            </div>
            <div className="p-3.5 bg-slate-50/80 rounded-2xl border border-slate-100 text-center">
              <div className="text-2xl font-black text-amber-600">100%</div>
              <div className="text-[11px] text-slate-500 font-bold">Двойная верификация</div>
            </div>
          </div>
        </section>

        {/* 3. ОНБОРДИНГ: КАК УСТРОЕН ПУТЬ ХРАНИТЕЛЯ (4 ШАГА) */}
        <section id="how-it-works" className="space-y-6">
          <div className="text-center max-w-2xl mx-auto space-y-2">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blue-100 text-[#1E50A0] text-xs font-black">
              <Layers className="w-3.5 h-3.5" />
              <span>ОНБОРДИНГ ДЛЯ НОВИЧКА</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-black text-[#0C2340]">
              Как устроен путь Хранителя за 4 шага
            </h2>
            <p className="text-xs sm:text-sm text-slate-500">
              Простой и увлекательный процесс сохранения культурного наследия
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 pt-2">
            {/* Step 1 */}
            <div className="bg-white rounded-3xl border border-slate-200/90 p-6 space-y-3 shadow-sm relative">
              <div className="w-10 h-10 rounded-2xl bg-[#0C2340] text-amber-400 font-black text-base flex items-center justify-center shadow-sm">
                1
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Задание дня и эпос</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Получайте ежедневные фразы из культурного корпуса («Урал-батыр», Мустай Карим, живая речь).
              </p>
            </div>

            {/* Step 2 */}
            <div className="bg-white rounded-3xl border border-slate-200/90 p-6 space-y-3 shadow-sm relative">
              <div className="w-10 h-10 rounded-2xl bg-[#1E50A0] text-white font-black text-base flex items-center justify-center shadow-sm">
                2
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Аутентичный башкирский TTS</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Слушайте чистое произношение специфических башкирских букв (ө, ү, ғ, ҡ, ң, ҙ, ҫ, һ) с кураем.
              </p>
            </div>

            {/* Step 3 */}
            <div className="bg-white rounded-3xl border border-slate-200/90 p-6 space-y-3 shadow-sm relative">
              <div className="w-10 h-10 rounded-2xl bg-emerald-600 text-white font-black text-base flex items-center justify-center shadow-sm">
                3
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Краудсорсинг и верификация</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Подтверждайте переводы или предлагайте правки, получая очки опыта (XP) и валюту «Ҡурай».
              </p>
            </div>

            {/* Step 4 */}
            <div className="bg-white rounded-3xl border border-slate-200/90 p-6 space-y-3 shadow-sm relative">
              <div className="w-10 h-10 rounded-2xl bg-amber-500 text-white font-black text-base flex items-center justify-center shadow-sm">
                4
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Лавка Хранителя и Школа</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Покупайте золотые рамки и титулы «Аҡһаҡал», а класс объединяется в экспедицию за первое место в РБ!
              </p>
            </div>
          </div>
        </section>

        {/* 4. БЛОК СРАВНЕНИЯ: ВСЕ ПЛЮСЫ И МИНУСЫ (ТРАДИЦИОННОЕ vs ЙӘНЛӘНДЕР) */}
        <section id="comparison" className="space-y-6">
          <div className="text-center max-w-2xl mx-auto space-y-2">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blue-100 text-[#1E50A0] text-xs font-black">
              <BookmarkCheck className="w-3.5 h-3.5" />
              <span>ОЗНАКОМИТЕЛЬНЫЙ АНАЛИЗ ПЛЮСОВ И МИНУСОВ</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-black text-[#0C2340]">
              Почему старые методы не работают, и в чём сила «Йәнләндер»
            </h2>
            <p className="text-xs sm:text-sm text-slate-500">
              Честное сравнение классического изучения с нашей интерактивной экосистемой
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
            {/* КАРТОЧКА МИНУСОВ (Традиционный подход) */}
            <div className="bg-red-50/60 rounded-3xl border border-red-200/80 p-6 sm:p-8 space-y-4 shadow-sm relative">
              <div className="flex items-center gap-3 pb-3 border-b border-red-200/60">
                <div className="w-11 h-11 rounded-2xl bg-red-100 text-red-600 flex items-center justify-center shrink-0">
                  <XCircle className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="text-base sm:text-lg font-black text-red-950">
                    Обычные методы и словари
                  </h3>
                  <p className="text-xs text-red-700 font-medium">
                    Минусы, из-за которых школьники теряют интерес
                  </p>
                </div>
              </div>

              <ul className="space-y-3.5 text-xs sm:text-sm text-slate-700">
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-red-200 text-red-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✕</span>
                  <span><strong>Сухая зубрежка:</strong> Бумажные словари и статичные таблицы не удерживают внимание современных детей дольше 10 минут.</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-red-200 text-red-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✕</span>
                  <span><strong>Нет правильной озвучки:</strong> Стандартные онлайн-переводчики пропускают или коверкают специфичные буквы (ө, ү, ғ, ҡ, ң, ҙ, ҫ, һ).</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-red-200 text-red-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✕</span>
                  <span><strong>Изолированность ученика:</strong> Занятия в одиночку без наград, уровней и соревновательного азарта быстро забрасываются.</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-red-200 text-red-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✕</span>
                  <span><strong>Галлюцинации нейросетей:</strong> Общедоступные LLM часто искажают башкирский контекст, выдавая некорректный машинный перевод.</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-red-200 text-red-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✕</span>
                  <span><strong>Школы не вовлечены:</strong> Нет платформы для дружеского соперничества между классами и лицеями разных районов Республики.</span>
                </li>
              </ul>

              <div className="pt-2">
                <div className="text-[11px] font-bold text-red-800 bg-red-100/80 p-3 rounded-2xl text-center">
                  Итог: низкая вовлеченность, забывание родного языка молодым поколением
                </div>
              </div>
            </div>

            {/* КАРТОЧКА ПЛЮСОВ (Платформа ЙӘНЛӘНДЕР) */}
            <div className="bg-emerald-50/60 rounded-3xl border border-emerald-200/90 p-6 sm:p-8 space-y-4 shadow-sm relative ring-1 ring-emerald-500/20">
              <div className="flex items-center gap-3 pb-3 border-b border-emerald-200/60">
                <div className="w-11 h-11 rounded-2xl bg-emerald-100 text-emerald-700 flex items-center justify-center shrink-0">
                  <CheckCircle2 className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="text-base sm:text-lg font-black text-emerald-950 flex items-center gap-2">
                    <span>Платформа «ЙӘНЛӘНДЕР»</span>
                    <span className="text-[10px] bg-emerald-600 text-white font-black px-2 py-0.5 rounded-full">РЕШЕНИЕ ХАКАТОНА</span>
                  </h3>
                  <p className="text-xs text-emerald-700 font-medium">
                    Плюсы и инновационные игровые механики проекта
                  </p>
                </div>
              </div>

              <ul className="space-y-3.5 text-xs sm:text-sm text-slate-800">
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-emerald-200 text-emerald-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✓</span>
                  <span><strong>Геймификация с первых секунд:</strong> XP за каждое слово, уровни Хранителя, ежедневные квесты и амулеты защиты стрика.</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-emerald-200 text-emerald-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✓</span>
                  <span><strong>Аутентичный башкирский TTS:</strong> Синтез речи с фонетической адаптацией редких букв и мелодичным колокольчиком курая.</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-emerald-200 text-emerald-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✓</span>
                  <span><strong>«Лавка Хранителя»:</strong> Экономика валюты «Ҡурай» — покупка золотых и сапфировых рамок профиля, титулов «Аҡһаҡал».</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-emerald-200 text-emerald-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✓</span>
                  <span><strong>Командные экспедиции классов:</strong> Совместный прогресс школ недели, битва гимназий в турнирной таблице.</span>
                </li>
                <li className="flex items-start gap-2.5">
                  <span className="w-5 h-5 rounded-full bg-emerald-200 text-emerald-800 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">✓</span>
                  <span><strong>Краудсорсинг + Админка:</strong> Народная верификация спорных текстов с утверждением экспертами и живой аналитикой Recharts.</span>
                </li>
              </ul>

              <div className="pt-2">
                <button
                  type="button"
                  onClick={() => handleProceed('student_timur@yanlander.ru')}
                  className="w-full py-3 rounded-2xl bg-emerald-600 hover:bg-emerald-700 text-white font-extrabold text-xs sm:text-sm transition-colors shadow-sm flex items-center justify-center gap-2 cursor-pointer"
                >
                  <span>Оценить в действии: Открыть дашборд</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </section>

        {/* 5. ИНТЕРАКТИВНОЕ ПРЕВЬЮ: ПОПРОБУЙТЕ БАШКИРСКУЮ ОЗВУЧКУ ПРЯМО СЕЙЧАС */}
        <section className="bg-white rounded-3xl border border-slate-200/90 p-6 sm:p-8 shadow-sm space-y-4">
          <div className="flex items-center gap-2">
            <span className="w-1.5 h-5 bg-[#1E50A0] rounded-full inline-block" />
            <h2 className="text-lg sm:text-xl font-bold text-[#0C2340]">
              Интерактивное превью: послушайте башкирское произношение
            </h2>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center">
            <div className="lg:col-span-8 space-y-3">
              <div className="flex items-center gap-3">
                <div className="text-xl sm:text-2xl font-black text-[#0C2340] font-serif">
                  «{currentFragment.text_ba}»
                </div>
                <AudioSpeakerButton text={currentFragment.text_ba} size="md" />
              </div>
              <div className="text-sm text-slate-600 font-medium">
                «{primaryTranslation}»
              </div>
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#EEF2FF] text-[#1E50A0] text-xs font-bold">
                <span>{currentFragment.source_title || 'Башкирский эпос "Урал-батыр"'}</span>
              </div>
            </div>

            <div className="lg:col-span-4 flex justify-end">
              <button
                type="button"
                onClick={() => handleProceed('student_timur@yanlander.ru')}
                className="w-full sm:w-auto px-6 py-3 rounded-full bg-[#0C2340] hover:bg-[#1A4480] text-white text-xs sm:text-sm font-bold shadow-md transition-all flex items-center justify-center gap-2 cursor-pointer"
              >
                <span>Перейти в Дашборд к заданиям</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </section>

        {/* 6. КЛЮЧЕВЫЕ ВОЗМОЖНОСТИ ЭКОСИСТЕМЫ */}
        <section id="features" className="space-y-6">
          <div className="flex items-center gap-2">
            <span className="w-1.5 h-5 bg-[#1E50A0] rounded-full inline-block" />
            <h2 className="text-xl sm:text-2xl font-black text-[#0C2340]">
              Все модули платформы «ЙӘНЛӘНДЕР»
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            <div className="p-6 bg-white rounded-3xl border border-slate-200/90 shadow-sm space-y-2.5">
              <div className="w-10 h-10 rounded-2xl bg-blue-50 text-[#1E50A0] flex items-center justify-center font-bold">
                <Volume2 className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Синтез башкирской речи (TTS)</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Фонетическая нормализация специфических букв, перелив курая и анимированный звуковой эквалайзер.
              </p>
            </div>

            <div className="p-6 bg-white rounded-3xl border border-slate-200/90 shadow-sm space-y-2.5">
              <div className="w-10 h-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
                <ShoppingBag className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">«Лавка Хранителя»</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Кастомизация профиля за монеты «Ҡурай»: солнечные и неоновые рамки аватара, титулы мудреца степей «Аҡһаҡал».
              </p>
            </div>

            <div className="p-6 bg-white rounded-3xl border border-slate-200/90 shadow-sm space-y-2.5">
              <div className="w-10 h-10 rounded-2xl bg-indigo-50 text-[#0C2340] flex items-center justify-center font-bold">
                <Users2 className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Школьные экспедиции</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Класс объединяется ради общей цели недели (100 переводов) и соревнуется в командном рейтинге школ.
              </p>
            </div>

            <div className="p-6 bg-white rounded-3xl border border-slate-200/90 shadow-sm space-y-2.5">
              <div className="w-10 h-10 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold">
                <BookOpen className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Живой интерактивный словарь</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Двусторонние карточки со звуком, фильтрация по сложности, контекстные примеры и персональное избранное.
              </p>
            </div>

            <div className="p-6 bg-white rounded-3xl border border-slate-200/90 shadow-sm space-y-2.5">
              <div className="w-10 h-10 rounded-2xl bg-purple-50 text-purple-600 flex items-center justify-center font-bold">
                <BarChart3 className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Аналитика оцифровки языка</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Графики Recharts за 7/30 дней, распределение по 22 тематическим категориям корпуса и вклад школ.
              </p>
            </div>

            <div className="p-6 bg-white rounded-3xl border border-slate-200/90 shadow-sm space-y-2.5">
              <div className="w-10 h-10 rounded-2xl bg-sky-50 text-sky-600 flex items-center justify-center font-bold">
                <Globe className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-sm text-[#0C2340]">Двуязычие RU / BA</h3>
              <p className="text-xs text-slate-500 leading-relaxed">
                Мгновенное переключение интерфейса на башкирский язык одним кликом (Төп бит, Хәзинә, Һүҙлек, Дәрәжә).
              </p>
            </div>
          </div>
        </section>

        {/* 7. BIG BOTTOM BANNER: ПРИСТУПИТЬ К РАБОТЕ */}
        <section className="bg-gradient-to-r from-[#0C2340] via-[#123561] to-[#1E50A0] rounded-3xl p-8 sm:p-12 text-white shadow-xl relative overflow-hidden">
          <div className="absolute right-0 top-0 bottom-0 w-1/3 opacity-10 flex items-center justify-center pointer-events-none">
            <BashkirOrnament className="w-80 h-80 text-white" />
          </div>

          <div className="relative z-10 max-w-2xl space-y-4">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/10 backdrop-blur-sm text-amber-300 text-xs font-bold">
              <Sparkles className="w-3.5 h-3.5" />
              <span>ГОТОВЫ СТАТЬ ХРАНИТЕЛЕМ БАШКИРСКОГО ЯЗЫКА?</span>
            </div>

            <h2 className="text-2xl sm:text-4xl font-black tracking-tight leading-tight">
              Откройте дашборд и начните путь прямо сейчас!
            </h2>

            <p className="text-sm text-slate-200 leading-relaxed">
              Ежедневные квесты, Лавка Хранителя, школьные битвы и культурное наследие ждут вас.
            </p>

            <div className="pt-2 flex flex-wrap items-center gap-4">
              <button
                type="button"
                onClick={() => handleProceed('student_timur@yanlander.ru')}
                disabled={isStarting}
                className="px-8 py-4 rounded-full bg-white text-[#0C2340] hover:bg-slate-100 text-sm sm:text-base font-black shadow-lg hover:shadow-xl transition-all flex items-center gap-3 transform hover:-translate-y-0.5 group cursor-pointer"
              >
                <Zap className="w-5 h-5 text-amber-500 fill-amber-500 group-hover:scale-110 transition-transform" />
                <span>{isStarting ? 'Загрузка...' : 'Приступить к изучению ➔'}</span>
              </button>

              <Link
                to="/login"
                className="px-6 py-4 rounded-full bg-white/10 hover:bg-white/20 text-white text-xs sm:text-sm font-bold border border-white/20 transition-colors"
              >
                Войти под своим логином
              </Link>
            </div>
          </div>
        </section>
      </main>

      {/* 8. CLEAN STANDALONE FOOTER */}
      <footer className="border-t border-slate-200/80 bg-white py-8 px-4 sm:px-8 text-center text-xs text-slate-500 space-y-2">
        <div className="flex items-center justify-center gap-2 font-bold text-[#0C2340]">
          <BashkirOrnament className="w-4 h-4 text-[#0C2340]" />
          <span>«ЙӘНЛӘНДЕР» — Хранители башкирского языка</span>
        </div>
        <p>© 2026 Хакатон Кибер-102. Разработано для сохранения и развития языка Республики Башкортостан.</p>
      </footer>
    </div>
  );
};