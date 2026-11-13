import React, { createContext, useContext, useState, useEffect } from 'react';

export type Language = 'ru' | 'ba';

interface Translations {
  [key: string]: {
    ru: string;
    ba: string;
  };
}

export const translations: Translations = {
  // Navigation
  nav_home: { ru: 'Главная', ba: 'Төп бит' },
  nav_tasks: { ru: 'Обучение', ba: 'Уҡыу' },
  nav_treasury: { ru: 'Сокровищница', ba: 'Хәзинә' },
  nav_leaderboard: { ru: 'Рейтинг', ba: 'Рейтинг' },
  nav_achievements: { ru: 'Достижения', ba: 'Уңыштар' },
  nav_team: { ru: 'Команда', ba: 'Команда' },
  nav_analytics: { ru: 'Аналитика', ba: 'Аналитика' },
  nav_dictionary: { ru: 'Словарь', ba: 'Һүҙлек' },
  nav_shop: { ru: 'Лавка Хранителя', ba: 'Һаҡлаусы кибете' },
  nav_profile: { ru: 'Профиль', ba: 'Профиль' },
  nav_admin: { ru: 'Модерация', ba: 'Модерация' },

  // Auth & General
  login: { ru: 'Войти', ba: 'Инеү' },
  register: { ru: 'Регистрация', ba: 'Теркәлеү' },
  logout: { ru: 'Выйти', ba: 'Сығыу' },
  search_placeholder: { ru: 'Поиск слов и переводов...', ba: 'Һүҙҙәрҙе һәм тәржемәләрҙе эҙләү...' },
  level: { ru: 'Уровень', ba: 'Дәрәжә' },
  streak_days: { ru: 'дн. ударно', ba: 'көн рәттән' },
  quray_currency: { ru: 'Ҡурай', ba: 'Ҡурай' },
  daily_task: { ru: 'Ежедневное задание', ba: 'Көн заданиеһы' },
  confirm: { ru: 'Подтвердить', ba: 'Раҫларға' },
  propose: { ru: 'Предложить исправление', ba: 'Төҙәтеү тәҡдим итергә' },
  listen_speech: { ru: 'Слушать произношение', ba: 'Әйтелеште тыңлау' },
};

interface LanguageContextType {
  lang: Language;
  setLang: (lang: Language) => void;
  toggleLang: () => void;
  t: (key: string) => string;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [lang, setLangState] = useState<Language>(() => {
    return (localStorage.getItem('yanlander_lang') as Language) || 'ru';
  });

  const setLang = (newLang: Language) => {
    setLangState(newLang);
    localStorage.setItem('yanlander_lang', newLang);
  };

  const toggleLang = () => {
    setLang(lang === 'ru' ? 'ba' : 'ru');
  };

  const t = (key: string): string => {
    if (translations[key]) {
      return translations[key][lang] || translations[key].ru;
    }
    return key;
  };

  return (
    <LanguageContext.Provider value={{ lang, setLang, toggleLang, t }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useLang = () => {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLang must be used within a LanguageProvider');
  }
  return context;
};
