import type { PageResult } from "../../types/api";
import type { Customer } from "../../types/customer";
import type { CustomerDetail } from "../../types/customerDetail";
import { apiGet } from "./client";
import type { ConversationMessageDto, CustomerDetailDto, CustomerDto, PageDto } from "./dto";
import { mapConversationMessage, mapCustomer, mapCustomerDetail, mapPage } from "./mappers";

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
