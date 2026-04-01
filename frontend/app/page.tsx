"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import {
  fetchCurrentUser,
  type UserResponse,
} from "@/lib/auth-api";
import { clearAuthToken, getAuthToken } from "@/lib/auth-token";

type HomeState =
  | { status: "guest" }
  | { status: "loading" }
  | { status: "ready"; user: UserResponse }
  | { status: "error"; message: string };

type PostDraft = {
  title: string;
  body: string;
  imageFile: File | null;
  contactEmail: string;
  contactVisibility: "public" | "approval_required" | "private";
};

const initialDraft: PostDraft = {
  title: "",
  body: "",
  imageFile: null,
  contactEmail: "",
  contactVisibility: "approval_required",
};

export default function HomePage() {
  const router = useRouter();
  const [state, setState] = useState<HomeState>({ status: "loading" });
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [draft, setDraft] = useState(initialDraft);
  const [postMessage, setPostMessage] = useState("");

  useEffect(() => {
    let active = true;

    async function loadUser() {
      const token = getAuthToken();
      if (!token) {
        if (active) {
          setState({ status: "guest" });
        }
        return;
      }

      const result = await fetchCurrentUser(token);
      if (!active) {
        return;
      }

      if (result.ok) {
        setState({ status: "ready", user: result.user });
        return;
      }

      if (result.status === 401) {
        clearAuthToken();
        setState({ status: "guest" });
        return;
      }

      setState({
        status: "error",
        message: result.message,
      });
    }

    loadUser();

    return () => {
      active = false;
    };
  }, []);

  function openPostModal() {
    setPostMessage("");
    if (state.status === "ready") {
      setDraft({
        title: "",
        body: "",
        imageFile: null,
        contactEmail: state.user.email,
        contactVisibility: state.user.email_visibility,
      });
    } else {
      setDraft(initialDraft);
    }
    setIsModalOpen(true);
  }

  function closePostModal() {
    setIsModalOpen(false);
    setPostMessage("");
  }

  function handlePostSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPostMessage(
      draft.imageFile
        ? `投稿 API はこれから実装します。選択中の画像: ${draft.imageFile.name}`
        : "投稿 API はこれから実装します。フォーム UI だけ先に用意しました。",
    );
    setDraft(initialDraft);
  }

  function handleLoginPrompt() {
    startTransition(() => {
      router.push("/login?redirect=/");
    });
  }

  return (
    <main className="relative flex min-h-screen flex-col overflow-hidden px-6 py-8 sm:px-10">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(255,236,183,0.92),_transparent_32%),radial-gradient(circle_at_bottom_right,_rgba(124,166,255,0.24),_transparent_26%)]" />

      <header className="relative mx-auto flex w-full max-w-6xl items-center justify-between gap-4">
        <Link className="brand-mark" href="/">
          MERUTOMO
        </Link>

        <div className="flex items-center gap-3">
          {state.status === "ready" ? (
            <Link className="profile-chip" href={`/${state.user.username}`}>
              {state.user.display_name}
            </Link>
          ) : (
            <>
              <Link className="retro-button retro-button-secondary" href="/login">
                Login
              </Link>
              <Link className="retro-button" href="/register">
                Register
              </Link>
            </>
          )}
        </div>
      </header>

      <section className="relative mx-auto mt-10 grid w-full max-w-6xl gap-8 lg:grid-cols-[1.1fr_0.9fr]">
        <div className="space-y-6">
          <div className="retro-panel p-7 sm:p-9">
            <p className="eyebrow">MERUTOMO BOARD</p>
            <h1 className="mt-3 text-4xl leading-[0.98] font-semibold tracking-[-0.06em] text-[var(--ink)] sm:text-6xl">
              届く前の余白を
              <br />
              掲示板から始める。
            </h1>
            <p className="mt-4 max-w-2xl text-base leading-8 text-[var(--muted)]">
              `/` はログイン後に最初に戻ってくるメイン画面です。投稿ボタンから募集文の作成に入れます。
              ログインしていない状態でも閲覧はでき、投稿時だけログイン導線へ誘導します。
            </p>

            <div className="mt-6 flex flex-wrap gap-3">
              <button className="retro-button" onClick={openPostModal} type="button">
                ＋ 投稿する
              </button>
              {state.status === "ready" ? (
                <Link className="retro-button retro-button-secondary" href={`/${state.user.username}`}>
                  マイページへ
                </Link>
              ) : (
                <Link className="retro-button retro-button-secondary" href="/login">
                  ログインする
                </Link>
              )}
            </div>

            {state.status === "error" ? (
              <p className="status-banner status-error mt-5">{state.message}</p>
            ) : null}
          </div>

          <div className="grid gap-4">
            {samplePosts.map((post) => (
              <article className="retro-panel p-6" key={post.username}>
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                      @{post.username}
                    </p>
                    <h2 className="mt-2 text-2xl font-semibold tracking-[-0.04em] text-[var(--ink)]">
                      {post.title}
                    </h2>
                  </div>
                  <span className="rounded-full border-2 border-[var(--line)] px-3 py-1 text-xs uppercase tracking-[0.18em] text-[var(--muted)]">
                    {post.visibilityLabel}
                  </span>
                </div>
                <p className="mt-4 text-sm leading-7 text-[var(--muted)]">{post.body}</p>
              </article>
            ))}
          </div>
        </div>

        <aside className="space-y-4">
          <div className="retro-panel p-6">
            <p className="eyebrow">Session</p>
            {state.status === "ready" ? (
              <div className="mt-3 space-y-3">
                <p className="text-3xl font-semibold tracking-[-0.05em] text-[var(--ink)]">
                  {state.user.display_name}
                </p>
                <p className="text-sm leading-7 text-[var(--muted)]">
                  username: @{state.user.username}
                  <br />
                  email visibility: {formatEmailVisibility(state.user.email_visibility)}
                </p>
              </div>
            ) : state.status === "loading" ? (
              <p className="mt-3 text-sm leading-7 text-[var(--muted)]">
                ログイン状態を確認しています...
              </p>
            ) : (
              <p className="mt-3 text-sm leading-7 text-[var(--muted)]">
                まだログインしていません。投稿やメール公開設定の確認にはログインが必要です。
              </p>
            )}
          </div>

          <div className="retro-panel p-6">
            <p className="eyebrow">Posting Flow</p>
            <ol className="mt-4 space-y-3 text-sm leading-7 text-[var(--muted)]">
              <li>1. `+` ボタンから投稿フォームを開く</li>
              <li>2. ログイン済みならそのまま入力</li>
              <li>3. 未ログインなら login へ誘導</li>
              <li>4. 次に投稿 API を接続する</li>
            </ol>
          </div>
        </aside>
      </section>

      <button
        aria-label="投稿する"
        className="floating-post-button"
        onClick={openPostModal}
        type="button"
      >
        +
      </button>

      {isModalOpen ? (
        <div className="modal-backdrop" role="presentation" onClick={closePostModal}>
          <div className="modal-card" role="dialog" aria-modal="true" onClick={(event) => event.stopPropagation()}>
            <div className="composer-header">
              <button className="composer-header-button" onClick={closePostModal} type="button">
                キャンセル
              </button>
              <h2 className="composer-title">新規スレッド</h2>
              <div className="composer-header-icons" aria-hidden="true">
                <span className="composer-icon">⧉</span>
                <span className="composer-icon">⋯</span>
              </div>
            </div>

            {state.status !== "ready" ? (
              <div className="composer-body">
                <div className="composer-author-row">
                  <div className="composer-avatar">?</div>
                  <div>
                    <p className="composer-handle">guest</p>
                    <p className="composer-subline">今なにしてる？</p>
                  </div>
                </div>
                <p className="status-banner status-info">
                  投稿するには login してください。ログイン後は `/` に戻ってそのまま投稿できます。
                </p>
                <div className="composer-footer">
                  <p className="composer-reply-option">↔ 返信オプション</p>
                  <button className="composer-submit-button" onClick={handleLoginPrompt} type="button">
                    Login へ進む
                  </button>
                </div>
              </div>
            ) : (
              <form className="composer-body" onSubmit={handlePostSubmit}>
                <div className="composer-author-row">
                  <div className="composer-avatar">
                    {state.user.display_name.slice(0, 1)}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="composer-handle">@{state.user.username}</p>
                      <span className="composer-topic-hint">トピックを追加</span>
                    </div>
                    <p className="composer-subline">今なにしてる？</p>
                  </div>
                </div>

                <label className="composer-field">
                  <input
                    className="composer-title-input"
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        title: event.target.value,
                      }))
                    }
                    placeholder="今夜、ゆっくり文通できる人"
                    value={draft.title}
                  />
                </label>

                <label className="composer-field">
                  <textarea
                    className="composer-textarea"
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        body: event.target.value,
                      }))
                    }
                    placeholder="自己紹介や、話したいことを書いてください。"
                    value={draft.body}
                  />
                </label>

                <div className="composer-toolbar" aria-hidden="true">
                  <span className="composer-tool">◫</span>
                  <span className="composer-tool">GIF</span>
                  <span className="composer-tool">☺</span>
                  <span className="composer-tool">≡</span>
                  <span className="composer-tool">🗒</span>
                  <span className="composer-tool">⌖</span>
                </div>

                <label className="composer-field">
                  <span className="composer-field-label">画像</span>
                  <label className="composer-file-picker">
                    <input
                      accept="image/*"
                      className="composer-file-input"
                      onChange={(event) =>
                        setDraft((current) => ({
                          ...current,
                          imageFile: event.target.files?.[0] ?? null,
                        }))
                      }
                      type="file"
                    />
                    <span className="composer-file-button">ファイルを選択</span>
                    <span className="composer-file-name">
                      {draft.imageFile ? draft.imageFile.name : "未選択"}
                    </span>
                  </label>
                </label>

                <div className="composer-contact-grid">
                  <label className="composer-field">
                    <span className="composer-field-label">メールアドレス</span>
                    <input
                      className="composer-secondary-input"
                      onChange={(event) =>
                        setDraft((current) => ({
                          ...current,
                          contactEmail: event.target.value,
                        }))
                      }
                      placeholder="contact@example.com"
                      type="email"
                      value={draft.contactEmail}
                    />
                  </label>

                  <label className="composer-field">
                    <span className="composer-field-label">公開範囲</span>
                    <select
                      className="composer-visibility-select"
                      onChange={(event) =>
                        setDraft((current) => ({
                          ...current,
                          contactVisibility: event.target.value as PostDraft["contactVisibility"],
                        }))
                      }
                      value={draft.contactVisibility}
                    >
                      <option value="approval_required">承認制</option>
                      <option value="public">公開</option>
                      <option value="private">非公開</option>
                    </select>
                  </label>
                </div>

                {postMessage ? <p className="status-banner status-info">{postMessage}</p> : null}

                <div className="composer-footer">
                  <p className="composer-reply-option">↔ 返信オプション</p>
                  <button className="composer-submit-button" type="submit">
                    投稿する
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      ) : null}
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

const samplePosts = [
  {
    username: "midnight_letter",
    title: "深夜2時の長文でも受け止めます",
    body: "映画、仕事、将来のこと。すぐ返さなくていいから、ちゃんと読みたい人とやり取りしたいです。",
    visibilityLabel: "approval",
  },
  {
    username: "retro_soda",
    title: "レトロ喫茶が好きな人",
    body: "便箋みたいな空気感の会話が好きです。写真付きの自己紹介をゆっくり交換できる相手を探しています。",
    visibilityLabel: "public",
  },
];
