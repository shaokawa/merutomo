"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import {
  fetchCurrentUser,
  logoutCurrentSession,
  type UserResponse,
} from "@/lib/auth-api";
import { clearAuthToken, getAuthToken } from "@/lib/auth-token";

type ProfileState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; user: UserResponse };

export default function UserPage() {
  const params = useParams<{ username: string }>();
  const router = useRouter();
  const [state, setState] = useState<ProfileState>({ status: "loading" });
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  useEffect(() => {
    const sessionToken = getAuthToken();
    if (!sessionToken) {
      startTransition(() => {
        router.replace(`/login?redirect=/${params.username}`);
      });
      return;
    }
    const token = sessionToken;

    let active = true;

    async function loadUser() {
      const result = await fetchCurrentUser(token);
      if (!active) {
        return;
      }

      if (!result.ok) {
        if (result.status === 401) {
          clearAuthToken();
          startTransition(() => {
            router.replace(`/login?redirect=/${params.username}`);
          });
          return;
        }

        setState({
          status: "error",
          message: result.message,
        });
        return;
      }

      if (result.user.username !== params.username) {
        startTransition(() => {
          router.replace(`/${result.user.username}`);
        });
        return;
      }

      setState({
        status: "ready",
        user: result.user,
      });
    }

    loadUser();

    return () => {
      active = false;
    };
  }, [params.username, router]);

  async function handleLogout() {
    const token = getAuthToken();
    if (token) {
      await logoutCurrentSession(token);
    }

    clearAuthToken();
    startTransition(() => {
      router.replace("/");
    });
  }

  return (
    <main className="page-shell">
      <section className="auth-card relative">
        <div className="absolute top-6 right-6">
          <button
            aria-label="メニューを開く"
            className="icon-button"
            onClick={() => setIsMenuOpen((current) => !current)}
            type="button"
          >
            ⋯
          </button>

          {isMenuOpen ? (
            <div className="menu-panel">
              <button className="menu-item danger-text" onClick={handleLogout} type="button">
                ログアウト
              </button>
            </div>
          ) : null}
        </div>

        {state.status === "loading" ? (
          <p className="status-banner status-info">プロフィールを読み込んでいます...</p>
        ) : null}

        {state.status === "error" ? (
          <p className="status-banner status-error">{state.message}</p>
        ) : null}

        {state.status === "ready" ? (
          <>
            <div className="space-y-3">
              <p className="eyebrow">MY PAGE</p>
              <h1 className="auth-title">{state.user.display_name}</h1>
              <p className="auth-copy">
                @{state.user.username}
                <br />
                email visibility: {formatEmailVisibility(state.user.email_visibility)}
              </p>
            </div>

            <div className="retro-panel mt-6 space-y-4 p-5">
              <p className="text-sm leading-7 text-[var(--muted)]">
                このページは `/{state.user.username}` です。現状はログイン中ユーザーのプロフィール確認用として使っています。
              </p>
              <p className="text-sm text-[var(--muted)]">
                display name: <span className="text-[var(--ink)]">{state.user.display_name}</span>
              </p>
              <p className="text-sm text-[var(--muted)]">
                email: <span className="text-[var(--ink)]">{state.user.email}</span>
              </p>
            </div>

            <div className="mt-6 flex flex-col gap-3 sm:flex-row">
              <Link className="retro-button retro-button-secondary" href="/">
                ホームへ戻る
              </Link>
              <button className="retro-button" onClick={handleLogout} type="button">
                ログアウト
              </button>
            </div>
          </>
        ) : null}
      </section>
    </main>
  );
}

function formatEmailVisibility(emailVisibility: UserResponse["email_visibility"]) {
  switch (emailVisibility) {
    case "public":
      return "公開";
    case "approval_required":
      return "承認制";
    case "private":
      return "非公開";
    default:
      return emailVisibility;
  }
}
