export type UserRole = 'admin' | 'learner'
export type UserStatus = 'active' | 'disabled'

export interface User {
  id: string
  email: string
  username: string
  role: UserRole
  status: UserStatus
  timezone: string
  last_login_at?: string
  created_at: string
}

export type InvitationStatus = 'pending' | 'accepted' | 'expired' | 'revoked'

export interface Invitation {
  id: string
  email: string
  status: InvitationStatus
  invited_by: string
  invited_by_username?: string
  expires_at: string
  accepted_at?: string
  revoked_at?: string
  created_at: string
  invitation_url?: string
  invitation_token?: string
}

export interface AuthResponse {
  access_token: string
  expires_in: number
  user: User
}

export interface ApiErrorDetail {
  field?: string
  message: string
}

export interface ApiErrorResponse {
  error: {
    code: string
    message: string
    details?: ApiErrorDetail[]
  }
}

// Curriculum Types

export type ContentStatus = 'draft' | 'published' | 'archived'
export type DifficultyLevel = 'beginner' | 'intermediate' | 'advanced'

export interface LearningPath {
  id: string
  title: string
  slug: string
  description?: string
  status: ContentStatus
  sort_order: number
  created_by?: string
  created_at: string
  updated_at: string
  published_at?: string
  total_lessons?: number
  completed_lessons?: number
  progress_percent?: number
}

export interface Course {
  id: string
  learning_path_id: string
  title: string
  slug: string
  description?: string
  status: ContentStatus
  sort_order: number
  created_at: string
  updated_at: string
  published_at?: string
  total_lessons?: number
  completed_lessons?: number
  progress_percent?: number
}

export interface Module {
  id: string
  course_id: string
  title: string
  slug: string
  description?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface Lesson {
  id: string
  module_id: string
  title: string
  slug: string
  summary?: string
  content: string
  difficulty: DifficultyLevel
  estimated_minutes: number
  status: ContentStatus
  sort_order: number
  created_at: string
  updated_at: string
  published_at?: string
}

export interface LessonSummary {
  id: string
  title: string
  slug: string
  summary?: string
  difficulty: DifficultyLevel
  estimated_minutes: number
  status: string
  sort_order: number
  is_completed: boolean
  is_locked: boolean
}

export interface ModuleWithLessons extends Module {
  lessons: LessonSummary[]
}

export interface CourseDetail extends Course {
  modules: ModuleWithLessons[]
  total_lessons: number
  completed_lessons: number
  progress_percent: number
}

export interface PrerequisiteInfo {
  id: string
  title: string
  slug: string
}

export interface LessonDetail extends Lesson {
  course_title: string
  course_slug: string
  module_title: string
  module_slug: string
  progress_status: string
  is_completed: boolean
  is_locked: boolean
  missing_prerequisites: PrerequisiteInfo[]
  prerequisites: PrerequisiteInfo[]
  prev_lesson_slug?: string
  next_lesson_slug?: string
}

export interface LessonProgress {
  id: string
  user_id: string
  lesson_id: string
  status: 'started' | 'completed'
  started_at: string
  completed_at?: string
  last_accessed_at: string
  updated_at: string
}


// Notes & Bookmarks Types

export interface Note {
  id: string
  user_id: string
  lesson_id?: string
  lesson_title?: string
  lesson_slug?: string
  title: string
  content: string
  pinned: boolean
  version: number
  tags: string[]
  created_at: string
  updated_at: string
}

export interface Tag {
  id: string
  name: string
  note_count: number
  created_at: string
}

export interface Bookmark {
  id: string
  user_id: string
  lesson_id: string
  lesson_title: string
  lesson_slug: string
  lesson_summary?: string
  lesson_difficulty: DifficultyLevel
  lesson_estimated_minutes: number
  course_title: string
  course_slug: string
  progress_status: string
  created_at: string
}
