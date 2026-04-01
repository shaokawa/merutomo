import { NextRequest, NextResponse } from "next/server";

import { proxyToBackend } from "@/lib/backend";

export async function POST(request: NextRequest) {
  try {
    return await proxyToBackend("/auth/logout", {
      method: "POST",
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
