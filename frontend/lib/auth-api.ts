export type UserResponse = {
  id: string;
  email: string;
  display_name: string;
  username: string;
  profile_text: string;
  avatar_url: string;
  email_visibility: "public" | "approval_required" | "private";
};

export type AuthResponse = {
  token: string;
  user: UserResponse;
  needs_email_confirmation?: boolean;
};

export type RegisterRequest = {
  email: string;
  password: string;
  display_name: string;
  username: string;
  email_visibility: "public" | "approval_required" | "private";
};

export type LoginRequest = {
  email: string;
  password: string;
};

type ErrorPayload = {
  error?: string;
  details?: string;
};

type MeResponse = {
  user: UserResponse;
};

export function readErrorMessage(payload: unknown) {
  const candidate = (payload || {}) as ErrorPayload;
  return candidate.details || candidate.error || "予期しないエラーが発生しました。";
}

export function isAuthResponse(payload: unknown): payload is AuthResponse {
  if (!payload || typeof payload !== "object") {
    return false;
  }

  return "user" in payload;
}

export function hasUser(payload: unknown): payload is { user: UserResponse } {
  if (!payload || typeof payload !== "object") {
    return false;
  }

  return "user" in payload;
}

export async function fetchCurrentUser(token: string): Promise<
  | { ok: true; user: UserResponse }
  | { ok: false; status: number; message: string }
> {
  try {
    const response = await fetch("/api/auth/me", {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const payload = (await response.json()) as MeResponse | ErrorPayload;
    if (!response.ok || !hasUser(payload)) {
      return {
        ok: false,
        status: response.status,
        message: readErrorMessage(payload),
      };
    }

    return {
      ok: true,
      user: payload.user,
    };
  } catch {
    return {
      ok: false,
      status: 502,
      message: "backend への接続に失敗しました。",
    };
  }
}

export async function logoutCurrentSession(token: string) {
  try {
    await fetch("/api/auth/logout", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  } catch {
    return;
  }
}
