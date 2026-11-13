export interface UserStats {
  user_id: number;
  xp: number;
  level: number;
  level_title: string;
  next_level_xp: number;
  currency: number;
  current_streak: number;
  best_streak: number;
  streak_protected: boolean;
  total_tasks_completed: number;
  total_translations_confirmed: number;
  total_proposals_made: number;
  total_accepted_proposals: number;
}

export interface User {
  id: number;
  email: string;
  username: string;
  full_name?: string;
  avatar: string;
  role: 'student' | 'teacher' | 'expert' | 'admin';
  school_id?: number;
  school_name?: string;
  class_id?: number;
  class_name?: string;
  created_at: string;
  stats?: UserStats;
}

export interface AuthResponse {
  access_token: string;
  token_type: string;
  user: User;
}

export interface Translation {
  id: number;
  fragment_id: number;
  translation_text: string;
  status: 'pending' | 'confirmed' | 'expert_review' | 'rejected';
  created_by?: number;
  creator_username?: string;
  votes_count: number;
  reject_count: number;
  user_voted?: 'confirm' | 'reject' | null;
  created_at: string;
}

export interface Fragment {
  id: number;
  text_ba: string;
  source_title: string;
  source_document?: string;
  chapter?: string;
  audio_url?: string;
  category: string;
  difficulty: 'easy' | 'medium' | 'hard';
  translations: Translation[];
}

export interface Achievement {
  id: number;
  code: string;
  name: string;
  description: string;
  icon: string;
  reward_xp: number;
  reward_currency: number;
  unlocked: boolean;
  unlocked_at?: string;
}

export interface DailyQuest {
  id: number;
  code: string;
  name: string;
  description: string;
  target: number;
  reward_xp: number;
  reward_currency: number;
  progress: number;
  completed: boolean;
  claimed: boolean;
}

export interface WeeklyExpedition {
  id: number;
  title: string;
  description: string;
  target_count: number;
  current_count: number;
  reward_xp: number;
  reward_currency: number;
  progress_percentage: number;
}

export interface LeaderboardUser {
  rank: number;
  user_id: number;
  username: string;
  full_name?: string;
  avatar: string;
  level: number;
  level_title: string;
  xp: number;
  streak: number;
  confirmed_count: number;
  school_name?: string;
}

export interface TeamLeaderboardEntry {
  rank: number;
  team_id: number;
  name: string;
  emblem: string;
  members_count: number;
  total_xp: number;
  total_confirmed: number;
}

export interface SchoolLeaderboardEntry {
  rank: number;
  school_id: number;
  name: string;
  city: string;
  students_count: number;
  total_xp: number;
  total_confirmed: number;
}

export interface DictionaryEntry {
  id: number;
  word: string;
  translation: string;
  context?: string;
  is_favorite: boolean;
  difficulty: string;
  created_at: string;
}

export interface TaskSubmissionResult {
  success: boolean;
  xp_awarded: number;
  currency_awarded: number;
  new_level?: number;
  level_up: boolean;
  streak_updated: boolean;
  current_streak: number;
  achievement_unlocked?: Achievement;
  message: string;
}

export interface PublicStats {
  total_users: number;
  total_fragments: number;
  confirmed_translations: number;
  total_contributions: number;
  active_schools: number;
}

export interface ActivityPoint {
  date: string;
  count: number;
}

export interface SchoolOption {
  id: number;
  name: string;
  city: string;
  classes: { id: number; name: string }[];
}

export interface AdminReviewItem {
  id: number;
  fragment_id: number;
  text_ba: string;
  source_title: string;
  translation_id: number;
  translation_text: string;
  status: string;
  votes_count: number;
  reject_count: number;
  creator_username?: string;
  created_at: string;
}
