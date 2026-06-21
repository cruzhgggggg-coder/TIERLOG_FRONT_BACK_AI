export type Role = 'student' | 'lecturer';

export type StudentProfile = {
  id: number;
  user_id: number;
  lecturer_id: number;
  nim: string;
  name: string;
  prodi: string;
  thesis_title: string;
};

export type LecturerProfile = {
  id: number;
  user_id: number;
  nip: string;
  name: string;
  faculty: string;
  keahlian: string;
  ai_constraints?: string;
};

export type User = {
  id: number;
  name: string;
  email: string;
  role: Role;
  openai_key?: string;
  gemini_key?: string;
  anthropic_key?: string;
  nvidia_key?: string;
  groq_key?: string;
  preferred_model?: string;
  is_gateway_active?: boolean;
  student?: StudentProfile;
  lecturer?: LecturerProfile;
};

export type FeedbackItem = {
  id: number;
  consultation_log_id: number;
  content: string;
  category: 'Major' | 'Minor';
  status: 'Pending' | 'Fixed' | 'Validated' | 'Rejected';
  comments?: FeedbackComment[];
};

export type FeedbackComment = {
  id: number;
  feedback_item_id: number;
  sender_id: number;
  sender_role: 'student' | 'lecturer';
  content: string;
  created_at: string;
};

export type RevisionAnnotation = {
  id: number;
  consultation_log_id: number;
  filename: string;
  file_type: 'image' | 'docx';
  extracted_text: string;
  created_at: string;
};

export type ConsultationLog = {
  id: number;
  student_id: number;
  audio_filename: string;
  transcript_filename: string;
  transcript_text: string;
  paper_filename: string;
  created_at: string;
  feedback_items: FeedbackItem[];
  revision_annotations?: RevisionAnnotation[];
  student?: StudentProfile & {
    user?: User;
    lecturer?: LecturerProfile;
  };
};

export type DashboardStats = {
  total_consultations: number;
  total_feedback: number;
  pending_feedback: number;
  major_feedback: number;
  completion_rate: number;
  draft_count: number;
  lecturer_name?: string;
  student_count?: number;
  validation_queue?: number;
  upcoming_quests: FeedbackItem[];
};

export type SessionInfo = {
  session_number: number;
  log_id: number;
  paper_filename: string;
  created_at: string;
  feedback_count: number;
  pending_count: number;
};

export type ModelCapability = 'text' | 'vision' | 'multimodal' | 'json' | 'streaming';

export type ModelInfo = {
  id: string;
  provider: string;
  capabilities: ModelCapability[];
  context_window: number;
  is_available: boolean;
};
