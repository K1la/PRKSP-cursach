"use client";

import { useState } from "react";
import { Activity } from "lucide-react";

import { getHealth } from "@/lib/api";
import { Button } from "@/components/ui/button";

type Status = "idle" | "loading" | "ok" | "error";

export function ApiStatusButton() {
  const [status, setStatus] = useState<Status>("idle");
  const [message, setMessage] = useState("Проверить API");

  async function handleCheck() {
    setStatus("loading");
    setMessage("Проверка...");
    try {
      const health = await getHealth();
      setStatus("ok");
      setMessage(`API ${health.status}, DB ${health.db}`);
    } catch {
      setStatus("error");
      setMessage("API недоступен");
    }
  }

  return (
    <Button
      className={status === "error" ? "bg-red-600 hover:bg-red-700" : undefined}
      onClick={handleCheck}
      disabled={status === "loading"}
    >
      <Activity className="mr-2 size-4" aria-hidden="true" />
      {message}
    </Button>
  );
}
