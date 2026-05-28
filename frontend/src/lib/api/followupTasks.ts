import type { PageResult } from "../../types/api";
import type { FollowupTask } from "../../types/followup";
import { apiGet } from "./client";
import type { FollowupTaskDto, PageDto } from "./dto";
import { mapFollowupTask, mapPage } from "./mappers";

export async function listPendingFollowupTasks(): Promise<PageResult<FollowupTask>> {
  const data = await apiGet<PageDto<FollowupTaskDto>>("/api/followup-tasks?status=pending&page=1&page_size=20");
  return mapPage(data, mapFollowupTask);
}
