"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { startTransition, useState } from "react";

import {
  type AuthResponse,
  type LoginRequest,
  isAuthResponse,
  readErrorMessage,
} from "@/lib/auth-api";
import { saveAuthToken } from "@/lib/auth-token";

const initialForm: LoginRequest = {
  email: "",
  password: "",
};

export default function LoginPage() {
  const router = useRouter();
  const [form, setForm] = useState(initialForm);
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(form),
      });

      const payload = (await response.json()) as AuthResponse | { error?: string; details?: string };
      if (!response.ok) {
        setError(readErrorMessage(payload));
        return;
      }

      if (!isAuthResponse(payload)) {
        setError("ログインレスポンスの形式が不正です。");
        return;
      }

      if (!payload.token) {
        setError("token が返ってきませんでした。認証フローを確認してください。");
        return;
      }

      saveAuthToken(payload.token);
      const redirectTo = readRedirect(new URLSearchParams(window.location.search).get("redirect"));
      startTransition(() => {
        router.push(redirectTo);
      });
    } catch {
      setError("ログインに失敗しました。backend が起動しているか確認してください。");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="page-shell">
      <section className="auth-card">
        <div className="space-y-3">
          <p className="eyebrow">Sign in</p>
          <h1 className="auth-title">ログイン</h1>
          <p className="auth-copy">
            ログイン成功時に access token を `localStorage` に保存し、welcome page から
            `/auth/me` を確認します。
          </p>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <label className="field">
            <span>Email</span>
            <input
              autoComplete="email"
              className="field-input"
              name="email"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  email: event.target.value,
                }))
              }
              required
              type="email"
              value={form.email}
            />
          </label>

          <label className="field">
            <span>Password</span>
            <input
              autoComplete="current-password"
              className="field-input"
              minLength={8}
              name="password"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  password: event.target.value,
                }))
              }
              required
              type="password"
              value={form.password}
            />
          </label>

          {error ? <p className="status-banner status-error">{error}</p> : null}

          <button className="retro-button w-full justify-center" disabled={isSubmitting} type="submit">
            {isSubmitting ? "ログイン中..." : "ログインする"}
          </button>
        </form>

        <p className="auth-footer">
          まだアカウントがないなら <Link href="/register">ユーザー登録へ</Link>
        </p>
      </section>
    </main>
  );
}

function readRedirect(value: string | null) {
  if (!value || !value.startsWith("/")) {
    return "/";
  }

  return value;
}
