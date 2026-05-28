import type { ApiResponseDto } from "./dto";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

export class ApiClientError extends Error {
  readonly code: string;
  readonly details?: Record<string, unknown>;

  constructor(code: string, message: string, details?: Record<string, unknown>) {
    super(message);
    this.name = "ApiClientError";
    this.code = code;
    this.details = details;
  }
}

export async function apiGet<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      ...init?.headers
    }
  });

  const body = (await response.json()) as ApiResponseDto<T>;
  if (!response.ok || body.error) {
    throw new ApiClientError(body.error?.code ?? "HTTP_ERROR", body.error?.message ?? "接口请求失败", body.error?.details);
  }
  if (body.data === null) {
    throw new ApiClientError("EMPTY_RESPONSE", "接口返回为空");
  }
  return body.data;
}
