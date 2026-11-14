import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowRight, BookOpen, Check, Globe2, Heart, Languages, Volume2 } from 'lucide-react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { useLang } from '../context/LanguageContext';
import { Fragment, PublicStats } from '../types';
import { BashkirOrnament } from '../components/ui/BashkirOrnament';
import { speakBashkir } from '../utils/speech';

export const LandingPage: React.FC = () => {
  const { user, quickLogin } = useAuth();
  const { lang, toggleLang } = useLang();
  const navigate = useNavigate();
  const [fragment, setFragment] = useState<Fragment | null>(null);
  const [stats, setStats] = useState<PublicStats | null>(null);
  const [isStarting, setIsStarting] = useState(false);

  useEffect(() => {
    api.getDailyFragment().then(setFragment).catch(() => {});
    api.getPublicStats().then(setStats).catch(() => {});
  }, []);

  const enterPlatform = async () => {
    setIsStarting(true);
    try {
      if (!user) await quickLogin('student_timur@yanlander.ru');
      navigate('/dashboard');
    } catch {
      navigate('/dashboard');
    } finally {
      setIsStarting(false);
    }
  };

  const text = fragment?.text_ba || 'Тыуған тел — халыҡтың рухи байлығы.';
  const translation = fragment?.translations?.[0]?.translation_text || 'Родной язык — духовное богатство народа.';

  return (
    <div className="yan-landing">
      <header className="yan-landing-header">
        <Link to="/" className="yan-brand"><span className="yan-brand-mark"><BashkirOrnament /></span><span><strong>ЙӘНЛӘНДЕР</strong><small>Хранители башкирского языка</small></span></Link>
        <nav className="yan-landing-nav"><a href="#how">Как это работает</a><a href="#community">Сообщество</a></nav>
        <div className="yan-header-actions">
          <button className="yan-language-switch" type="button" onClick={toggleLang}><Languages size={16} /> {lang === 'ru' ? 'РУС' : 'БАШ'}</button>
          {user ? <button type="button" onClick={() => navigate('/dashboard')} className="yan-header-login">В кабинет <ArrowRight size={15} /></button> : <Link to="/login" className="yan-header-login">Войти <ArrowRight size={15} /></Link>}
        </div>
      </header>

      <main>
        <section className="yan-landing-hero">
          <div className="yan-hero-copy">
            <span className="yan-eyebrow"><Heart size={15} /> Язык, который мы храним вместе</span>
            <h1>Каждое слово<br /><em>сохраняет историю.</em></h1>
            <p>Помогай собирать башкирские тексты, сверять переводы и передавать язык дальше. Начни с одного фрагмента — твой вклад уже важен.</p>
            <div className="yan-hero-actions"><button type="button" onClick={enterPlatform} disabled={isStarting} className="yan-button">{isStarting ? 'Открываем…' : 'Попробовать платформу'} <ArrowRight size={17} /></button><a href="#how" className="yan-text-link">Узнать, как это работает <ArrowRight size={15} /></a></div>
            <div className="yan-hero-stats"><span><strong>{stats?.total_fragments.toLocaleString('ru-RU') || '—'}</strong> фрагментов</span><span><strong>{stats?.total_users.toLocaleString('ru-RU') || '—'}</strong> хранителей</span></div>
          </div>
          <div className="yan-hero-art" aria-label="Башкирские буквы и орнамент">
            <div className="yan-orbit yan-orbit-one" /><div className="yan-orbit yan-orbit-two" />
            <div className="yan-letter yan-letter-a">Ә</div><div className="yan-letter yan-letter-o">Ө</div><div className="yan-letter yan-letter-k">Ҡ</div>
            <div className="yan-hero-ornament"><BashkirOrnament /></div>
            <span className="yan-art-caption">Һүҙ — халыҡтың хазинаһы</span>
          </div>
        </section>

        <section className="yan-demo-section" aria-label="Пример фрагмента">
          <div className="yan-demo-label"><span className="yan-eyebrow">ПОПРОБУЙ ПРЯМО СЕЙЧАС</span><span className="yan-source-tag">Фрагмент дня</span></div>
          <div className="yan-demo-content"><div><span className="yan-language-label">БАШҠОРТСА</span><p className="yan-demo-bashkir">{text}</p><p className="yan-demo-translation">{translation}</p></div><button type="button" className="yan-audio-button" onClick={() => speakBashkir(text)} aria-label="Послушать фрагмент"><Volume2 size={20} /></button></div>
          <div className="yan-demo-footer"><span>{fragment?.source_title || 'Башкирская народная мудрость'}</span><button type="button" onClick={enterPlatform} className="yan-text-link">Помочь проверить перевод <ArrowRight size={15} /></button></div>
        </section>

        <section id="how" className="yan-how-section">
          <div className="yan-section-intro"><span className="yan-eyebrow">НЕБОЛЬШИЕ ШАГИ. ОБЩЕЕ НАСЛЕДИЕ.</span><h2>Язык живёт в общении</h2><p>Присоединяйся к сообществу, которое бережно собирает и проверяет башкирские тексты.</p></div>
          <div className="yan-steps">
            <article><span className="yan-step-number">01</span><BookOpen size={23} /><h3>Читай и слушай</h3><p>Знакомься с короткими фрагментами и живым звучанием башкирской речи.</p></article>
            <article><span className="yan-step-number">02</span><Check size={23} /><h3>Сверяй перевод</h3><p>Предлагай свой вариант или помогай сообществу проверить смысл.</p></article>
            <article><span className="yan-step-number">03</span><Globe2 size={23} /><h3>Сохраняй вместе</h3><p>Каждый подтверждённый текст пополняет общий языковой архив.</p></article>
          </div>
        </section>

        <section id="community" className="yan-community-cta"><div><span className="yan-eyebrow">БЕРЕЖНО СОХРАНЯЕМ КУЛЬТУРУ</span><h2>Твоё первое слово может стать частью общего наследия.</h2></div><button type="button" onClick={enterPlatform} disabled={isStarting} className="yan-button">Начать <ArrowRight size={17} /></button></section>
      </main>
      <footer className="yan-footer"><Link to="/" className="yan-brand"><span className="yan-brand-mark"><BashkirOrnament /></span><span><strong>ЙӘНЛӘНДЕР</strong><small>Сохраняй язык. Создавай наследие.</small></span></Link><span>Башкирский язык — наше общее наследие.</span></footer>
    </div>
  );
};

export default LandingPage;
