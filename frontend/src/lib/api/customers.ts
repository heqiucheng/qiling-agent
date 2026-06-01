import type { PageResult } from "../../types/api";
import type { Customer } from "../../types/customer";
import type { CustomerDetail } from "../../types/customerDetail";
import type { LongTermMemory, MemoryFactCorrectionResult, MemoryFactStatusResult } from "../../types/memory";
import { apiGet, apiPost } from "./client";
import type { ConversationMessageDto, CustomerDetailDto, CustomerDto, LongTermMemoryDto, MemoryFactCorrectionResultDto, MemoryFactStatusResultDto, PageDto } from "./dto";
import { mapConversationMessage, mapCustomer, mapCustomerDetail, mapLongTermMemory, mapMemoryFactCorrectionResult, mapMemoryFactStatusResult, mapPage } from "./mappers";

export async function listCustomers(): Promise<PageResult<Customer>> {
  const data = await apiGet<PageDto<CustomerDto>>("/api/customers?page=1&page_size=20");
  return mapPage(data, mapCustomer);
}

export async function getCustomerDetail(customerId: string): Promise<CustomerDetail> {
  const [detail, conversations] = await Promise.all([
    apiGet<CustomerDetailDto>(`/api/customers/${customerId}`),
    apiGet<PageDto<ConversationMessageDto>>(`/api/customers/${customerId}/conversations?page=1&page_size=50`)
  ]);
  return mapCustomerDetail(detail, conversations.items.map(mapConversationMessage));
}

export async function getCustomerLongTermMemory(customerId: string): Promise<LongTermMemory> {
  const data = await apiGet<LongTermMemoryDto>(`/api/customers/${customerId}/long-term-memory`);
  return mapLongTermMemory(data);
}

export async function rejectLongTermMemoryFact(customerId: string, factId: string, reason: string): Promise<MemoryFactStatusResult> {
  const data = await apiPost<MemoryFactStatusResultDto, Record<string, unknown>>(`/api/customers/${customerId}/long-term-memory/facts/${factId}/reject`, {
    reason
  });
  return mapMemoryFactStatusResult(data);
}

export async function correctLongTermMemoryFact(customerId: string, factId: string, payload: {
  category: string;
  key: string;
  value: string;
  confidence: number;
  reason: string;
}): Promise<MemoryFactCorrectionResult> {
  const data = await apiPost<MemoryFactCorrectionResultDto, Record<string, unknown>>(`/api/customers/${customerId}/long-term-memory/facts/${factId}/correct`, payload);
  return mapMemoryFactCorrectionResult(data);
}
