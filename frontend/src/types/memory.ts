import type { Customer } from "./customer";

export type MemoryFactStatus = "active" | "superseded" | "rejected";

export interface LongTermMemoryFact {
  id: string;
  customerId: string;
  category: string;
  key: string;
  value: string;
  confidence: number;
  sourceType: string;
  sourceId: string;
  status: MemoryFactStatus;
  createdAt: string;
  updatedAt: string;
}

export interface LongTermMemory {
  customer: Customer;
  facts: LongTermMemoryFact[];
  promptContext: string;
  builtAt: string;
}

export interface MemoryFactStatusResult {
  factId: string;
  status: MemoryFactStatus;
}

export interface MemoryFactCorrectionResult {
  oldFactId: string;
  oldStatus: MemoryFactStatus;
  newFact: LongTermMemoryFact;
}
