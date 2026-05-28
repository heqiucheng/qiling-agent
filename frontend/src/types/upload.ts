export interface UploadConversationResult {
  uploadId: string;
  status: "uploaded" | "parsed" | "needs_confirmation" | "confirmed" | "failed";
  parsedCustomer: {
    name: string;
    ownerName: string;
  };
  messageCount: number;
  warnings: string[];
  nextAction: string;
}

export interface ConfirmUploadResult {
  customerId: string;
  conversationId: string;
  agentRunId: string;
  followupTaskId: string;
  status: "confirmed";
}
