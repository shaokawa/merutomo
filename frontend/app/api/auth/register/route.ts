import { NextRequest, NextResponse } from "next/server";

import { proxyToBackend } from "@/lib/backend";

export async function POST(request: NextRequest) {
  try {
    return await proxyToBackend("/auth/register", {
      method: "POST",
      body: await request.text(),
      headers: {
        "Content-Type": "application/json",
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
