import type { ConfirmUploadResult, UploadConversationResult } from "../../types/upload";
import { apiPost } from "./client";
import type { ConfirmUploadResultDto, UploadConversationResultDto } from "./dto";
import { mapConfirmUploadResult, mapUploadConversationResult } from "./mappers";

export async function uploadConversation(content: string): Promise<UploadConversationResult> {
  const data = await apiPost<UploadConversationResultDto, Record<string, unknown>>("/api/uploads/conversations", {
    source_type: "pasted_text",
    content,
    owner_id: "usr_001"
  });
  return mapUploadConversationResult(data);
}

export async function confirmUpload(uploadId: string, customerName: string): Promise<ConfirmUploadResult> {
  const data = await apiPost<ConfirmUploadResultDto, Record<string, unknown>>(`/api/uploads/${uploadId}/confirm`, {
    customer_name: customerName,
    owner_id: "usr_001"
  });
  return mapConfirmUploadResult(data);
}
