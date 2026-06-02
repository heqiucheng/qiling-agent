import type { ApiResponseDto } from "./dto";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";
const DEFAULT_USER_ID = "usr_001";
const DEFAULT_ROLE = "sales";

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
      ...authHeaders(),
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

export async function apiPost<TResponse, TRequest extends Record<string, unknown>>(path: string, payload: TRequest): Promise<TResponse> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...authHeaders()
    },
    body: JSON.stringify(payload)
  });

  const body = (await response.json()) as ApiResponseDto<TResponse>;
  if (!response.ok || body.error) {
    throw new ApiClientError(body.error?.code ?? "HTTP_ERROR", body.error?.message ?? "接口请求失败", body.error?.details);
  }
  if (body.data === null) {
    throw new ApiClientError("EMPTY_RESPONSE", "接口返回为空");
  }
  return body.data;
}

export async function apiGetText(path: string, init?: RequestInit): Promise<string> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: "text/markdown,text/plain,*/*",
      ...authHeaders(),
      ...init?.headers
    }
  });

  if (!response.ok) {
    throw new ApiClientError("HTTP_ERROR", await response.text());
  }
  return response.text();
}

function authHeaders(): Record<string, string> {
  const role = window.localStorage.getItem("qiling_mock_role") ?? DEFAULT_ROLE;
  return {
    "X-Qiling-User-ID": role === "manager" ? "mgr_001" : DEFAULT_USER_ID,
    "X-Qiling-Role": role
  };
}
