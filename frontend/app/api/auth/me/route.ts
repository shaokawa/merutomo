import { NextRequest, NextResponse } from "next/server";

import { proxyToBackend } from "@/lib/backend";

export async function GET(request: NextRequest) {
  try {
    return await proxyToBackend("/auth/me", {
      headers: {
        Authorization: request.headers.get("authorization") ?? "",
      },
    });
  } catch {
    return NextResponse.json(
      {
        error: "backend unavailable",
      },
      { status: 502 },
    );
  }
}
