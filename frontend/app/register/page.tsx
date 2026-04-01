"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { startTransition, useState } from "react";

import {
  type AuthResponse,
  type RegisterRequest,
  isAuthResponse,
  readErrorMessage,
} from "@/lib/auth-api";
import { saveAuthToken } from "@/lib/auth-token";

const initialForm: RegisterRequest = {
  email: "",
  password: "",
  display_name: "",
  username: "",
  email_visibility: "approval_required",
};

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState(initialForm);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccess("");
    setIsSubmitting(true);

    try {
      const response = await fetch("/api/auth/register", {
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
        setError("登録レスポンスの形式が不正です。");
        return;
      }

      if (payload.token) {
        saveAuthToken(payload.token);
        const redirectTo = readRedirect(new URLSearchParams(window.location.search).get("redirect"));
        startTransition(() => {
          router.push(redirectTo);
        });
        return;
      }

      setSuccess(
        "登録できました。確認メールを開いたあと、ログイン画面からサインインしてください。",
      );
      setForm(initialForm);
    } catch {
      setError("登録に失敗しました。backend が起動しているか確認してください。");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="page-shell">
      <section className="auth-card">
        <div className="space-y-3">
          <p className="eyebrow">Create account</p>
          <h1 className="auth-title">ユーザー登録</h1>
          <p className="auth-copy">
            `display_name` は welcome page の表示名です。`username` は公開IDです。
            メールアドレスの公開範囲は登録時点で選べます。
          </p>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <label className="field">
            <span>Display Name</span>
            <input
              autoComplete="nickname"
              className="field-input"
              name="display_name"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  display_name: event.target.value,
                }))
              }
              required
              value={form.display_name}
            />
          </label>

          <label className="field">
            <span>Username</span>
            <input
              autoCapitalize="off"
              autoComplete="username"
              className="field-input"
              name="username"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  username: event.target.value,
                }))
              }
              pattern="[a-z0-9_]{3,30}"
              placeholder="taiga_testuser"
              required
              value={form.username}
            />
          </label>

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
            <span>Email Visibility</span>
            <select
              className="field-input"
              name="email_visibility"
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  email_visibility: event.target.value as RegisterRequest["email_visibility"],
                }))
              }
              value={form.email_visibility}
            >
              <option value="approval_required">approval required</option>
              <option value="public">public</option>
              <option value="private">private</option>
            </select>
          </label>

          <label className="field">
            <span>Password</span>
            <input
              autoComplete="new-password"
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
          {success ? <p className="status-banner status-success">{success}</p> : null}

          <button className="retro-button w-full justify-center" disabled={isSubmitting} type="submit">
            {isSubmitting ? "登録中..." : "登録する"}
          </button>
        </form>

        <p className="auth-footer">
          すでに登録済みなら <Link href="/login">ログインへ</Link>
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
