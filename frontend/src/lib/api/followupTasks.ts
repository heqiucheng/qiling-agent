import type { PageResult } from "../../types/api";
import type { AgentRecommendation } from "../../types/agent";
import type { FollowupTask } from "../../types/followup";
import { apiGet, apiPost } from "./client";
import type { FollowupTaskDto, PageDto, RegenerateTaskResultDto, TaskCopyResultDto, TaskStatusResultDto, MarkWrongResultDto } from "./dto";
import { mapFollowupTask, mapPage, mapRecommendation } from "./mappers";

export async function listPendingFollowupTasks(): Promise<PageResult<FollowupTask>> {
  const data = await apiGet<PageDto<FollowupTaskDto>>("/api/followup-tasks?status=pending&page=1&page_size=20");
  return mapPage(data, mapFollowupTask);
}

export async function copyFollowupTask(taskId: string, copiedScript: string): Promise<TaskCopyResultDto> {
  return apiPost<TaskCopyResultDto, Record<string, unknown>>(`/api/followup-tasks/${taskId}/copy`, {
    copied_script: copiedScript,
    client_copied_at: new Date().toISOString()
  });
}

export async function skipFollowupTask(taskId: string): Promise<TaskStatusResultDto> {
  return apiPost<TaskStatusResultDto, Record<string, unknown>>(`/api/followup-tasks/${taskId}/skip`, {
    reason: "销售暂时跳过"
  });
}

export async function markFollowupTaskWrong(taskId: string): Promise<MarkWrongResultDto> {
  return apiPost<MarkWrongResultDto, Record<string, unknown>>(`/api/followup-tasks/${taskId}/mark-wrong`, {
    reason: "销售认为当前判断不准确",
    wrong_fields: ["recommended_action"]
  });
}

export async function regenerateFollowupTask(taskId: string): Promise<AgentRecommendation> {
  const data = await apiPost<RegenerateTaskResultDto, Record<string, unknown>>(`/api/followup-tasks/${taskId}/regenerate`, {
    instruction: "语气更自然一点，不要太像营销话术"
  });
  return mapRecommendation(data.recommendation);
}
