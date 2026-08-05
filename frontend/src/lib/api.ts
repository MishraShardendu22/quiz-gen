import { config } from "../config";

export interface ApiResponse<T> {
  code: number;
  success: boolean;
  message: string;
  data: T;
}

export interface Topic {
  id: string;
  name: string;
  status: string;
  document_count: number;
}

export interface Question {
  id: string;
  session_id: string;
  question: string;
  option_1: string;
  option_2: string;
  option_3: string;
  option_4: string;
  correct_answer: number;
  explanation: string;
  created_at: number;
}

export type SessionStatus = "pending" | "processing" | "completed" | "failed";

export interface Session {
  id: string;
  topic_id: string;
  status: SessionStatus;
  requested_count: number;
  generated_count: number;
  token_budget: number;
  tokens_used: number;
  created_at: number;
  updated_at: number;
  error?: string | null;
  questions?: Question[];
}

export interface GenerateRequest {
  topic_id: string;
  requested_count: number;
  token_budget: number;
}

export interface GenerateResponse {
  session_id: string;
  status: string;
}

export interface UsageBreakdown {
  session_id: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  estimated_cost: number;
}

export interface UsageReport {
  total_prompt_tokens: number;
  total_completion_tokens: number;
  total_tokens: number;
  estimated_cost: number;
  sessions: UsageBreakdown[];
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${config.backendBaseUrl}${path}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options?.headers || {}),
    },
  });

  const payload = await response.json();

  if (!response.ok || payload.success === false) {
    const errorMsg = payload.error || payload.message || `HTTP Error ${response.status}`;
    throw new Error(errorMsg);
  }

  return payload.data as T;
}

export const api = {
  getTopics: (): Promise<Topic[]> => request<Topic[]>("/topics"),

  getSessions: (): Promise<Session[]> => request<Session[]>("/sessions"),

  getSession: (id: string): Promise<Session> => request<Session>(`/sessions/${id}`),

  createGenerate: (data: GenerateRequest): Promise<GenerateResponse> =>
    request<GenerateResponse>("/generate", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  retrySession: (id: string, tokenBudget: number): Promise<GenerateResponse> =>
    request<GenerateResponse>(`/sessions/${id}/retry`, {
      method: "POST",
      body: JSON.stringify({ token_budget: tokenBudget }),
    }),

  getUsage: (): Promise<UsageReport> => request<UsageReport>("/usage"),
};
