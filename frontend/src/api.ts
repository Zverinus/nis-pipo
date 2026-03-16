const API_URL = import.meta.env.VITE_API_URL || '';

function getToken(): string | null {
  const t = localStorage.getItem('token');
  if (!t) return null;
  const trimmed = t.trim();
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    try {
      return JSON.parse(trimmed) as string;
    } catch {
      return trimmed;
    }
  }
  return trimmed;
}

export function decodeTokenPayload(token: string): { email?: string } | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const payload = atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(payload) as { email?: string };
  } catch {
    return null;
  }
}

export async function api<T>(
  path: string,
  options: RequestInit & { token?: boolean } = {}
): Promise<T> {
  const { token = true, method, ...init } = options;
  const headers: Record<string, string> = {
    ...(init.headers as Record<string, string>),
  };
  if (method && method !== 'GET') {
    headers['Content-Type'] = 'application/json';
  }
  const t = getToken();
  if (token && t) headers['Authorization'] = `Bearer ${t}`;
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    method,
    headers,
    cache: 'no-store',
  });
  if (!res.ok) {
    const text = await res.text();
    if (res.status === 401 && token && getToken()) {
      localStorage.removeItem('token');
      const reason = res.headers.get('X-Auth-Reason');
      const params = reason === 'token_expired' ? '?session_expired=1' : '';
      window.location.replace(`/login${params}`);
    }
    throw new Error(text || `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const auth = {
  me: () => api<{ id: string; email: string }>('/api/auth/me'),
  register: (email: string, password: string) =>
    api<{ id: string; email: string }>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
      token: false,
    }),
  login: (email: string, password: string) =>
    api<{ token: string }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
      token: false,
    }),
};

export const meetings = {
  list: () => api<Meeting[] | null>('/api/meetings').then((data) => data ?? []),
  get: (id: string) => api<Meeting>(`/api/meetings/${id}`, { token: false }),
  create: (data: CreateMeetingInput) =>
    api<Meeting>('/api/meetings', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  results: (id: string) =>
    api<SlotResult[] | null>(`/api/meetings/${id}/results`).then((data) => data ?? []),
  finalize: (id: string, finalSlotIndex: number) =>
    api<void>(`/api/meetings/${id}/finalize`, {
      method: 'PUT',
      body: JSON.stringify({ final_slot_index: finalSlotIndex }),
    }),
};

export const participants = {
  create: (meetingId: string, displayName: string) =>
    api<{ id: string }>(
      `/api/meetings/${meetingId}/participants`,
      {
        method: 'POST',
        body: JSON.stringify({ display_name: displayName }),
        token: false,
      }
    ),
  setSlots: (meetingId: string, participantId: string, slotIndexes: number[]) =>
    api<void>(`/api/meetings/${meetingId}/participants/${participantId}/slots`, {
      method: 'PUT',
      body: JSON.stringify({ slot_indexes: slotIndexes }),
      token: false,
    }),
};

export interface Meeting {
  id: string;
  owner_id: string;
  title: string;
  description: string;
  date_start: string;
  date_end: string;
  slot_minutes: number;
  status: string;
  final_slot_index?: number;
}

export interface SlotResult {
  slot_index: number;
  count: number;
  participant_names?: string[];
}

export interface CreateMeetingInput {
  title: string;
  description: string;
  date_start: string;
  date_end: string;
  slot_minutes: number;
}
