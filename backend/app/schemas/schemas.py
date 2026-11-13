from pydantic import BaseModel, EmailStr, Field
from typing import Optional, List, Dict, Any
from datetime import datetime, date

# ----------------- Auth & User Schemas -----------------
class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"
    user: "UserOut"

class TokenPayload(BaseModel):
    sub: Optional[str] = None
    exp: Optional[int] = None

class UserLogin(BaseModel):
    email: str
    password: str

class UserRegister(BaseModel):
    email: EmailStr
    username: str
    password: str
    full_name: Optional[str] = None
    role: str = "student"
    school_id: Optional[int] = None
    class_id: Optional[int] = None

class UserUpdate(BaseModel):
    full_name: Optional[str] = None
    avatar: Optional[str] = None
    school_id: Optional[int] = None
    class_id: Optional[int] = None

class UserStatsOut(BaseModel):
    user_id: int
    xp: int
    level: int
    level_title: str
    next_level_xp: int
    currency: int
    current_streak: int
    best_streak: int
    streak_protected: bool
    total_tasks_completed: int
    total_translations_confirmed: int
    total_proposals_made: int
    total_accepted_proposals: int

    class Config:
        from_attributes = True

class UserOut(BaseModel):
    id: int
    email: str
    username: str
    full_name: Optional[str] = None
    avatar: str
    role: str
    school_id: Optional[int] = None
    school_name: Optional[str] = None
    class_id: Optional[int] = None
    class_name: Optional[str] = None
    created_at: datetime
    stats: Optional[UserStatsOut] = None
    cosmetics: Optional[Dict[str, Any]] = None

    class Config:
        from_attributes = True

# ----------------- Gamification & Quests -----------------
class AchievementOut(BaseModel):
    id: int
    code: str
    name: str
    description: str
    icon: str
    reward_xp: int
    reward_currency: int
    unlocked: bool = False
    unlocked_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class DailyQuestOut(BaseModel):
    id: int
    code: str
    name: str
    description: str
    target: int
    reward_xp: int
    reward_currency: int
    progress: int = 0
    completed: bool = False
    claimed: bool = False

    class Config:
        from_attributes = True

class WeeklyExpeditionOut(BaseModel):
    id: int
    title: str
    description: str
    target_count: int
    current_count: int
    reward_xp: int
    reward_currency: int
    progress_percentage: float

    class Config:
        from_attributes = True

class LeaderboardUser(BaseModel):
    rank: int
    user_id: int
    username: str
    full_name: Optional[str] = None
    avatar: str
    level: int
    level_title: str
    xp: int
    streak: int
    confirmed_count: int
    school_name: Optional[str] = None

class TeamLeaderboardEntry(BaseModel):
    rank: int
    team_id: int
    name: str
    emblem: str
    members_count: int
    total_xp: int
    total_confirmed: int

class SchoolLeaderboardEntry(BaseModel):
    rank: int
    school_id: int
    name: str
    city: str
    students_count: int
    total_xp: int
    total_confirmed: int

# ----------------- Fragments & Translations -----------------
class TranslationOut(BaseModel):
    id: int
    fragment_id: int
    translation_text: str
    status: str
    created_by: Optional[int] = None
    creator_username: Optional[str] = None
    votes_count: int
    reject_count: int
    user_voted: Optional[str] = None  # None, "confirm", "reject"
    created_at: datetime

    class Config:
        from_attributes = True

class FragmentOut(BaseModel):
    id: int
    text_ba: str
    source_title: str
    source_document: Optional[str] = None
    chapter: Optional[str] = None
    audio_url: Optional[str] = None
    category: str
    difficulty: str
    translations: List[TranslationOut] = []

    class Config:
        from_attributes = True

class TranslationCreate(BaseModel):
    fragment_id: int
    translation_text: str

class TranslationVoteRequest(BaseModel):
    vote_type: str = "confirm"  # "confirm" or "reject"

class TranslationHistoryOut(BaseModel):
    id: int
    translation_id: int
    previous_text: str
    new_text: str
    changed_by_username: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True

class TaskSubmissionResult(BaseModel):
    success: bool
    xp_awarded: int
    currency_awarded: int
    new_level: Optional[int] = None
    level_up: bool = False
    streak_updated: bool = False
    current_streak: int = 0
    achievement_unlocked: Optional[AchievementOut] = None
    message: str

# ----------------- Dictionary -----------------
class DictionaryEntryCreate(BaseModel):
    fragment_id: Optional[int] = None
    word: str
    translation: str
    context: Optional[str] = None
    difficulty: str = "medium"

class DictionaryEntryOut(BaseModel):
    id: int
    word: str
    translation: str
    context: Optional[str] = None
    is_favorite: bool
    difficulty: str
    created_at: datetime

    class Config:
        from_attributes = True

# ----------------- Admin & Export -----------------
class AdminReviewItem(BaseModel):
    id: int
    fragment_id: int
    text_ba: str
    source_title: str
    translation_id: int
    translation_text: str
    status: str
    votes_count: int
    reject_count: int
    creator_username: Optional[str] = None
    created_at: datetime

class FragmentCreateAdmin(BaseModel):
    text_ba: str
    source_title: str
    source_document: Optional[str] = None
    category: str = "Общие фразы"
    difficulty: str = "medium"
    initial_translation: Optional[str] = None

class SystemOverviewStats(BaseModel):
    total_users: int
    total_fragments: int
    total_translations: int
    confirmed_translations: int
    in_review_translations: int
    total_contributions: int
    active_schools_count: int
