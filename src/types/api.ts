export interface ApiMeta {
  requestId: string;
  timestamp: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiResponse<T> {
  data: T | null;
  error: ApiError | null;
  meta: ApiMeta;
}