import type { ReviewSummary } from "../../types/review";
import { apiGet } from "./client";
import type { ReviewSummaryDto } from "./dto";
import { mapReviewSummary } from "./mappers";

export async function getReviewSummary(): Promise<ReviewSummary> {
  const data = await apiGet<ReviewSummaryDto>("/api/review-reports/summary");
  return mapReviewSummary(data);
}
