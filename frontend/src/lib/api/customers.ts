import type { PageResult } from "../../types/api";
import type { Customer } from "../../types/customer";
import { apiGet } from "./client";
import type { CustomerDto, PageDto } from "./dto";
import { mapCustomer, mapPage } from "./mappers";

export async function listCustomers(): Promise<PageResult<Customer>> {
  const data = await apiGet<PageDto<CustomerDto>>("/api/customers?page=1&page_size=20");
  return mapPage(data, mapCustomer);
}
