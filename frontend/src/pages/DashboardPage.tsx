import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, BookOpen, Headphones, Leaf, Trophy, Volume2 } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { api } from '../api/client';
import { Fragment, PublicStats, LeaderboardUser, DictionaryEntry, DailyQuest } from '../types';
import { speakBashkir } from '../utils/speech';

export const DashboardPage: React.FC = () => {
  const { user, openAuthModal } = useAuth();
  const [stats, setStats] = useState<PublicStats | null>(null);
  const [dailyFragment, setDailyFragment] = useState<Fragment | null>(null);
  const [leaders, setLeaders] = useState<LeaderboardUser[]>([]);
  const [archiveWord, setArchiveWord] = useState<DictionaryEntry | null>(null);
  const [dailyQuests, setDailyQuests] = useState<DailyQuest[]>([]);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  useEffect(() => {
    api.getPublicStats().then(setStats).catch(() => {});
    api.getDailyFragment().then(setDailyFragment).catch(() => {});
    api.getLeaderboard('weekly', 'xp').then((data) => setLeaders(Array.isArray(data) ? data.slice(0, 3) : [])).catch(() => {});
    api.getDictionary().then((data) => setArchiveWord(Array.isArray(data) ? data[0] ?? null : null)).catch(() => {});
    api.getDailyQuests().then((data) => setDailyQuests(Array.isArray(data) ? data : [])).catch(() => {});
  }, []);

  const playAudio = () => {
    if (!dailyFragment?.text_ba) return;
    speakBashkir(dailyFragment.text_ba, {
      onStart: () => setIsPlayingAudio(true),
      onEnd: () => setIsPlayingAudio(false),
      onError: () => setIsPlayingAudio(false),
    });
  };

  const streak = user?.stats?.current_streak ?? 0;
  const quest = dailyQuests[0];

  return (
    <div className="yan-dashboard">
      <section className="yan-welcome">
        <div className="yan-welcome-copy">
          <span className="yan-eyebrow"><Leaf size={15} /> Язык живёт, пока им делятся</span>
          <h1>{user ? `С возвращением, ${user.full_name?.split(' ')[0] || 'хранитель'}!` : 'Сохраняй язык.\nСоздавай наследие.'}</h1>
          <p>Помоги сохранить башкирские слова и истории — один небольшой вклад за раз.</p>
          <Link to="/task" className="yan-button"><BookOpen size={17} /> Начать практику <ArrowRight size={16} /></Link>
        </div>
        <div className="yan-welcome-mark" aria-hidden="true"><span>Ә</span><i>Ө</i><b>Ҡ</b><em>Ғ</em></div>
        <div className="yan-welcome-note">Һүҙ — халыҡтың хазинаһы<br /><small>Слово — сокровище народа</small></div>
      </section>

      <section className="yan-fragment" aria-labelledby="fragment-title">
        <div className="yan-section-heading">
          <div><span className="yan-eyebrow">НЕБОЛЬШОЙ ШАГ К БОЛЬШОМУ ДЕЛУ</span><h2 id="fragment-title">Фрагмент дня</h2></div>
          <span className="yan-source-tag">{dailyFragment?.category || 'Пословица'}</span>
        </div>
        <div className="yan-quote">
          <div>
            <span className="yan-language-label">БАШҠОРТСА</span>
            <p>{dailyFragment?.text_ba || 'Тыуған тел — халыҡтың рухи байлығы.'}</p>
            <span className="yan-muted">{dailyFragment?.source_title || 'Башкирская народная мудрость'}</span>
          </div>
          <button type="button" className={`yan-audio-button${isPlayingAudio ? ' is-playing' : ''}`} onClick={playAudio} aria-label="Прослушать фрагмент" title="Прослушать фрагмент">
            {isPlayingAudio ? <Headphones size={20} /> : <Volume2 size={20} />}
          </button>
        </div>
        <div className="yan-translation"><span>ПО-РУССКИ</span><p>{dailyFragment?.translations?.[0]?.translation_text || 'Родной язык — духовное богатство народа.'}</p></div>
        <div className="yan-fragment-footer"><span>Проверь перевод или предложи свой вариант — сообщество поможет сверить смысл.</span><Link to="/task" className="yan-text-link">К заданию <ArrowRight size={16} /></Link></div>
      </section>

      <section className="yan-community-strip" aria-label="Сообщество платформы">
        <div className="yan-stat"><span>Фрагментов в корпусе</span><strong>{stats ? stats.total_fragments.toLocaleString('ru-RU') : '—'}</strong></div>
        <div className="yan-stat"><span>Хранителей языка</span><strong>{stats ? stats.total_users.toLocaleString('ru-RU') : '—'}</strong></div>
        <div className="yan-stat"><span>Твоя серия</span><strong>{user ? `${streak} дней` : 'Присоединяйся'}</strong></div>
      </section>

      <section className="yan-lower-grid">
        <div className="yan-lower-block">
          <div className="yan-section-heading compact"><div><span className="yan-eyebrow">СЛОВО ИЗ АРХИВА</span><h2>Язык в деталях</h2></div><Link to="/treasury" className="yan-icon-link" aria-label="Открыть сокровищницу"><ArrowRight size={17} /></Link></div>
          <div className="yan-word">{archiveWord?.word || 'ҡояш'}<span>{archiveWord?.translation || 'солнце'}</span></div>
          {archiveWord?.context && <p className="yan-muted yan-context">«{archiveWord.context}»</p>}
          <Link to="/treasury" className="yan-text-link">Сокровищница языка <ArrowRight size={15} /></Link>
        </div>

        <div className="yan-lower-block">
          <div className="yan-section-heading compact"><div><span className="yan-eyebrow">ВМЕСТЕ МЫ СИЛЬНЕЕ</span><h2>Лидеры недели</h2></div><Trophy className="yan-blue-icon" size={21} /></div>
          {leaders.length ? <ol className="yan-leaders">{leaders.map((leader, index) => <li key={leader.user_id || index}><span className="yan-rank">0{index + 1}</span><span>{leader.full_name || leader.username}</span><strong>{leader.xp.toLocaleString('ru-RU')} XP</strong></li>)}</ol> : <p className="yan-muted">Рейтинг сообщества появится здесь.</p>}
          <Link to="/leaderboard" className="yan-text-link">Весь рейтинг <ArrowRight size={15} /></Link>
        </div>

        <div className="yan-lower-block yan-quest-block">
          <span className="yan-eyebrow">ТВОЙ ВКЛАД СЕГОДНЯ</span>
          <h2>{quest?.name || 'Каждое слово имеет значение'}</h2>
          <p>{quest?.description || 'Проверь один перевод и помоги пополнить общий корпус.'}</p>
          {quest && <div className="yan-progress"><span style={{ width: `${Math.min(100, Math.round((quest.progress / Math.max(quest.target, 1)) * 100))}%` }} /></div>}
          <Link to="/task" className="yan-text-link">Продолжить <ArrowRight size={15} /></Link>
        </div>
      </section>

      {!user && <button className="yan-join-note" type="button" onClick={() => openAuthModal('/profile')}>Сохраняй историю вместе с нами <ArrowRight size={15} /></button>}
    </div>
  );
};

export default DashboardPage;
