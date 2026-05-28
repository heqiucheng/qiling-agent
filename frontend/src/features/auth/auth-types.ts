export type MockRole = "sales" | "manager";

export interface MockUser {
  id: string;
  name: string;
  role: MockRole;
  roleLabel: string;
}
