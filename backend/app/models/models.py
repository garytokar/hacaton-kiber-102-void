from datetime import datetime, date
from typing import Optional, List
from sqlalchemy import (
    String, Integer, Boolean, Text, ForeignKey, DateTime, Date, Float, UniqueConstraint, Index
)
from sqlalchemy.orm import Mapped, mapped_column, relationship
from app.core.database import Base

class School(Base):
    __tablename__ = "schools"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    city: Mapped[str] = mapped_column(String(100), default="Уфа")

    classes: Mapped[List["SchoolClass"]] = relationship("SchoolClass", back_populates="school", cascade="all, delete-orphan")
    users: Mapped[List["User"]] = relationship("User", back_populates="school")


class SchoolClass(Base):
    __tablename__ = "classes"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    school_id: Mapped[int] = mapped_column(Integer, ForeignKey("schools.id", ondelete="CASCADE"), nullable=False)
    name: Mapped[str] = mapped_column(String(50), nullable=False)
    grade_number: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
    letter: Mapped[Optional[str]] = mapped_column(String(10), nullable=True)
    invite_code: Mapped[Optional[str]] = mapped_column(String(50), unique=True, nullable=True)

    school: Mapped["School"] = relationship("School", back_populates="classes")
    users: Mapped[List["User"]] = relationship("User", back_populates="school_class")


class Team(Base):
    __tablename__ = "teams"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    emblem: Mapped[str] = mapped_column(String(100), default="shield")
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    members: Mapped[List["TeamMember"]] = relationship("TeamMember", back_populates="team", cascade="all, delete-orphan")


class TeamMember(Base):
    __tablename__ = "team_members"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    team_id: Mapped[int] = mapped_column(Integer, ForeignKey("teams.id", ondelete="CASCADE"), nullable=False)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    joined_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    team: Mapped["Team"] = relationship("Team", back_populates="members")
    user: Mapped["User"] = relationship("User", back_populates="team_memberships")

    __table_args__ = (UniqueConstraint("team_id", "user_id", name="uq_team_user"),)


class Language(Base):
    __tablename__ = "languages"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    code: Mapped[str] = mapped_column(String(10), unique=True, nullable=False)  # ba, ru, en
    name: Mapped[str] = mapped_column(String(50), nullable=False)


class LanguagePair(Base):
    __tablename__ = "language_pairs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    source_language_id: Mapped[int] = mapped_column(Integer, ForeignKey("languages.id"), nullable=False)
    target_language_id: Mapped[int] = mapped_column(Integer, ForeignKey("languages.id"), nullable=False)

    source_lang: Mapped["Language"] = relationship("Language", foreign_keys=[source_language_id])
    target_lang: Mapped["Language"] = relationship("Language", foreign_keys=[target_language_id])


class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True, nullable=False)
    username: Mapped[str] = mapped_column(String(100), unique=True, index=True, nullable=False)
    password_hash: Mapped[str] = mapped_column(String(255), nullable=False)
    full_name: Mapped[Optional[str]] = mapped_column(String(150), nullable=True)
    avatar: Mapped[str] = mapped_column(String(255), default="avatar_default")
    role: Mapped[str] = mapped_column(String(20), default="student")  # student, teacher, expert, admin
    preferred_lang_pair: Mapped[str] = mapped_column(String(20), default="ba-ru")
    
    school_id: Mapped[Optional[int]] = mapped_column(Integer, ForeignKey("schools.id", ondelete="SET NULL"), nullable=True)
    class_id: Mapped[Optional[int]] = mapped_column(Integer, ForeignKey("classes.id", ondelete="SET NULL"), nullable=True)
    
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    # Relationships
    school: Mapped[Optional["School"]] = relationship("School", back_populates="users")
    school_class: Mapped[Optional["SchoolClass"]] = relationship("SchoolClass", back_populates="users")
    stats: Mapped[Optional["UserStats"]] = relationship("UserStats", back_populates="user", uselist=False, cascade="all, delete-orphan")
    team_memberships: Mapped[List["TeamMember"]] = relationship("TeamMember", back_populates="user", cascade="all, delete-orphan")
    answers: Mapped[List["UserAnswer"]] = relationship("UserAnswer", back_populates="user", cascade="all, delete-orphan")
    dictionary_entries: Mapped[List["DictionaryEntry"]] = relationship("DictionaryEntry", back_populates="user", cascade="all, delete-orphan")
    achievements: Mapped[List["UserAchievement"]] = relationship("UserAchievement", back_populates="user", cascade="all, delete-orphan")
    quests: Mapped[List["UserQuest"]] = relationship("UserQuest", back_populates="user", cascade="all, delete-orphan")


class UserStats(Base):
    __tablename__ = "user_stats"

    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), primary_key=True)
    xp: Mapped[int] = mapped_column(Integer, default=0, index=True)
    level: Mapped[int] = mapped_column(Integer, default=1)
    currency: Mapped[int] = mapped_column(Integer, default=50)  # Ҡурай
    current_streak: Mapped[int] = mapped_column(Integer, default=0)
    best_streak: Mapped[int] = mapped_column(Integer, default=0)
    last_activity_date: Mapped[Optional[date]] = mapped_column(Date, nullable=True)
    streak_protected: Mapped[bool] = mapped_column(Boolean, default=False)
    
    # Quality metrics
    total_tasks_completed: Mapped[int] = mapped_column(Integer, default=0)
    total_translations_confirmed: Mapped[int] = mapped_column(Integer, default=0)
    total_proposals_made: Mapped[int] = mapped_column(Integer, default=0)
    total_accepted_proposals: Mapped[int] = mapped_column(Integer, default=0)

    user: Mapped["User"] = relationship("User", back_populates="stats")


class Fragment(Base):
    __tablename__ = "fragments"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    text_ba: Mapped[str] = mapped_column(Text, nullable=False)
    source_title: Mapped[str] = mapped_column(String(255), default="Народный фольклор")
    source_document: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    chapter: Mapped[Optional[str]] = mapped_column(String(100), nullable=True)
    audio_url: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    category: Mapped[str] = mapped_column(String(100), default="Пословицы и поговорки")
    difficulty: Mapped[str] = mapped_column(String(20), default="medium")  # easy, medium, hard
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    translations: Mapped[List["Translation"]] = relationship("Translation", back_populates="fragment", cascade="all, delete-orphan")


class Translation(Base):
    __tablename__ = "translations"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    fragment_id: Mapped[int] = mapped_column(Integer, ForeignKey("fragments.id", ondelete="CASCADE"), nullable=False, index=True)
    language_pair_id: Mapped[int] = mapped_column(Integer, ForeignKey("language_pairs.id"), default=1)
    translation_text: Mapped[str] = mapped_column(Text, nullable=False)
    status: Mapped[str] = mapped_column(String(30), default="pending", index=True)  # pending, confirmed, expert_review, rejected
    created_by: Mapped[Optional[int]] = mapped_column(Integer, ForeignKey("users.id", ondelete="SET NULL"), nullable=True)
    votes_count: Mapped[int] = mapped_column(Integer, default=0)
    reject_count: Mapped[int] = mapped_column(Integer, default=0)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    fragment: Mapped["Fragment"] = relationship("Fragment", back_populates="translations")
    creator: Mapped[Optional["User"]] = relationship("User")
    votes: Mapped[List["TranslationVote"]] = relationship("TranslationVote", back_populates="translation", cascade="all, delete-orphan")
    history: Mapped[List["TranslationHistory"]] = relationship("TranslationHistory", back_populates="translation", cascade="all, delete-orphan")


class TranslationVote(Base):
    __tablename__ = "translation_votes"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    translation_id: Mapped[int] = mapped_column(Integer, ForeignKey("translations.id", ondelete="CASCADE"), nullable=False)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    vote_type: Mapped[str] = mapped_column(String(20), default="confirm")  # confirm, reject
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    translation: Mapped["Translation"] = relationship("Translation", back_populates="votes")
    user: Mapped["User"] = relationship("User")

    __table_args__ = (UniqueConstraint("translation_id", "user_id", name="uq_translation_user_vote"),)


class TranslationHistory(Base):
    __tablename__ = "translation_history"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    translation_id: Mapped[int] = mapped_column(Integer, ForeignKey("translations.id", ondelete="CASCADE"), nullable=False)
    previous_text: Mapped[str] = mapped_column(Text, nullable=False)
    new_text: Mapped[str] = mapped_column(Text, nullable=False)
    changed_by: Mapped[Optional[int]] = mapped_column(Integer, ForeignKey("users.id", ondelete="SET NULL"), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    translation: Mapped["Translation"] = relationship("Translation", back_populates="history")


class UserAnswer(Base):
    __tablename__ = "user_answers"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False, index=True)
    fragment_id: Mapped[int] = mapped_column(Integer, ForeignKey("fragments.id", ondelete="CASCADE"), nullable=False)
    translation_id: Mapped[Optional[int]] = mapped_column(Integer, ForeignKey("translations.id", ondelete="SET NULL"), nullable=True)
    action: Mapped[str] = mapped_column(String(30), nullable=False)  # confirm, propose, reject
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow, index=True)

    user: Mapped["User"] = relationship("User", back_populates="answers")
    fragment: Mapped["Fragment"] = relationship("Fragment")


class Achievement(Base):
    __tablename__ = "achievements"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    code: Mapped[str] = mapped_column(String(50), unique=True, nullable=False)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    description: Mapped[str] = mapped_column(Text, nullable=False)
    icon: Mapped[str] = mapped_column(String(50), default="award")
    condition_type: Mapped[str] = mapped_column(String(50), nullable=False)  # tasks_count, streak_days, proposals_count, confirmed_count, xp_total
    condition_value: Mapped[int] = mapped_column(Integer, nullable=False)
    reward_xp: Mapped[int] = mapped_column(Integer, default=50)
    reward_currency: Mapped[int] = mapped_column(Integer, default=20)


class UserAchievement(Base):
    __tablename__ = "user_achievements"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    achievement_id: Mapped[int] = mapped_column(Integer, ForeignKey("achievements.id", ondelete="CASCADE"), nullable=False)
    unlocked_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    user: Mapped["User"] = relationship("User", back_populates="achievements")
    achievement: Mapped["Achievement"] = relationship("Achievement")

    __table_args__ = (UniqueConstraint("user_id", "achievement_id", name="uq_user_achievement"),)


class DailyQuest(Base):
    __tablename__ = "daily_quests"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    code: Mapped[str] = mapped_column(String(50), unique=True, nullable=False)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    description: Mapped[str] = mapped_column(Text, nullable=False)
    target: Mapped[int] = mapped_column(Integer, default=1)
    reward_xp: Mapped[int] = mapped_column(Integer, default=10)
    reward_currency: Mapped[int] = mapped_column(Integer, default=5)
    quest_type: Mapped[str] = mapped_column(String(50), default="confirm")  # confirm, verify, propose


class UserQuest(Base):
    __tablename__ = "user_quests"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    quest_id: Mapped[int] = mapped_column(Integer, ForeignKey("daily_quests.id", ondelete="CASCADE"), nullable=False)
    quest_date: Mapped[date] = mapped_column(Date, default=date.today)
    progress: Mapped[int] = mapped_column(Integer, default=0)
    completed: Mapped[bool] = mapped_column(Boolean, default=False)
    claimed: Mapped[bool] = mapped_column(Boolean, default=False)

    user: Mapped["User"] = relationship("User", back_populates="quests")
    quest: Mapped["DailyQuest"] = relationship("DailyQuest")

    __table_args__ = (UniqueConstraint("user_id", "quest_id", "quest_date", name="uq_user_quest_date"),)


class DictionaryEntry(Base):
    __tablename__ = "dictionary_entries"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False, index=True)
    fragment_id: Mapped[Optional[int]] = mapped_column(Integer, ForeignKey("fragments.id", ondelete="SET NULL"), nullable=True)
    word: Mapped[str] = mapped_column(String(150), nullable=False)
    translation: Mapped[str] = mapped_column(String(255), nullable=False)
    context: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    is_favorite: Mapped[bool] = mapped_column(Boolean, default=False)
    difficulty: Mapped[str] = mapped_column(String(20), default="medium")  # easy, medium, hard
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

    user: Mapped["User"] = relationship("User", back_populates="dictionary_entries")
    fragment: Mapped[Optional["Fragment"]] = relationship("Fragment")


class WeeklyExpedition(Base):
    __tablename__ = "weekly_expeditions"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    title: Mapped[str] = mapped_column(String(150), nullable=False)
    description: Mapped[str] = mapped_column(Text, nullable=False)
    target_count: Mapped[int] = mapped_column(Integer, default=100)
    current_count: Mapped[int] = mapped_column(Integer, default=0)
    reward_xp: Mapped[int] = mapped_column(Integer, default=50)
    reward_currency: Mapped[int] = mapped_column(Integer, default=30)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)


class Notification(Base):
    __tablename__ = "notifications"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False, index=True)
    title: Mapped[str] = mapped_column(String(150), nullable=False)
    message: Mapped[str] = mapped_column(Text, nullable=False)
    is_read: Mapped[bool] = mapped_column(Boolean, default=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
